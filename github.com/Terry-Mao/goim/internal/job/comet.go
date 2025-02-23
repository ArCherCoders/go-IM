package job

import (
	"context"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"sync/atomic"
	"time"

	"github.com/Terry-Mao/goim/api/comet"
	"github.com/Terry-Mao/goim/internal/job/conf"
	log "github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	// grpc options
	grpcKeepAliveTime    = time.Duration(10) * time.Second
	grpcKeepAliveTimeout = time.Duration(20) * time.Second
	grpcBackoffMaxDelay  = time.Duration(3) * time.Second
	grpcMaxSendMsgSize   = 1 << 24
	grpcMaxCallMsgSize   = 1 << 24
)

const (
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
)

func newCometClient(addr string) (comet.CometClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr,
		[]grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
			grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                grpcKeepAliveTime,    // 超过这个时长都没有活动的化 客户端就会ping服务端 这个最小值是10秒
				Timeout:             grpcKeepAliveTimeout, // 发送Ping后 客户端等待回复 超过这个时长没有收到ping的回复消息 则会断开连接 默认值是20秒
				PermitWithoutStream: true,                 // 即使没有活动流 也发送Ping
			}),
		}...,
	)
	if err != nil {
		return nil, err
	}
	return comet.NewCometClient(conn), err
}

// Comet is a comet.
type Comet struct {
	serverID      string
	client        comet.CometClient
	pushChan      []chan *comet.PushMsgReq
	roomChan      []chan *comet.BroadcastRoomReq
	broadcastChan chan *comet.BroadcastReq
	pushChanNum   uint64
	roomChanNum   uint64
	routineSize   uint64
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewComet new a comet.
func NewComet(in *model.Instance, c *conf.Comet) (*Comet, error) {
	cmt := &Comet{
		serverID:      in.Ip,
		pushChan:      make([]chan *comet.PushMsgReq, c.RoutineSize),       //32
		roomChan:      make([]chan *comet.BroadcastRoomReq, c.RoutineSize), //32
		broadcastChan: make(chan *comet.BroadcastReq, c.RoutineSize),       //32
		routineSize:   uint64(c.RoutineSize),                               //32
	}

	IP := in.Ip
	port := in.Port
	addr := fmt.Sprintf("%s:%d", IP, port)
	var err error
	if cmt.client, err = newCometClient(addr); err != nil {
		return nil, err
	}
	cmt.ctx, cmt.cancel = context.WithCancel(context.Background())

	for i := 0; i < c.RoutineSize; i++ {
		cmt.pushChan[i] = make(chan *comet.PushMsgReq, c.RoutineChan)
		cmt.roomChan[i] = make(chan *comet.BroadcastRoomReq, c.RoutineChan)
		go cmt.process(cmt.pushChan[i], cmt.roomChan[i], cmt.broadcastChan)
	}
	return cmt, nil
}

// Push push a user message.
func (c *Comet) Push(arg *comet.PushMsgReq) (err error) {
	idx := atomic.AddUint64(&c.pushChanNum, 1) % c.routineSize
	c.pushChan[idx] <- arg
	return
}

// BroadcastRoom broadcast a room message.
func (c *Comet) BroadcastRoom(arg *comet.BroadcastRoomReq) (err error) {
	idx := atomic.AddUint64(&c.roomChanNum, 1) % c.routineSize
	c.roomChan[idx] <- arg
	return
}

// Broadcast broadcast a message.
func (c *Comet) Broadcast(arg *comet.BroadcastReq) (err error) {
	c.broadcastChan <- arg
	return
}

func (c *Comet) process(pushChan chan *comet.PushMsgReq, roomChan chan *comet.BroadcastRoomReq, broadcastChan chan *comet.BroadcastReq) {
	for {
		select {
		case broadcastArg := <-broadcastChan:
			_, err := c.client.Broadcast(context.Background(), &comet.BroadcastReq{
				Proto:   broadcastArg.Proto,
				ProtoOp: broadcastArg.ProtoOp,
				Speed:   broadcastArg.Speed,
			})
			if err != nil {
				log.Errorf("c.client.Broadcast(%s, reply) serverId:%s error(%v)", broadcastArg, c.serverID, err)
			}
		case roomArg := <-roomChan:
			_, err := c.client.BroadcastRoom(context.Background(), &comet.BroadcastRoomReq{
				RoomID: roomArg.RoomID,
				Proto:  roomArg.Proto,
			})
			if err != nil {
				log.Errorf("c.client.BroadcastRoom(%s, reply) serverId:%s error(%v)", roomArg, c.serverID, err)
			}
		case pushArg := <-pushChan:
			_, err := c.client.PushMsg(context.Background(), &comet.PushMsgReq{
				Keys:    pushArg.Keys,
				Proto:   pushArg.Proto,
				ProtoOp: pushArg.ProtoOp,
			})
			if err != nil {
				log.Errorf("c.client.PushMsg(%s, reply) serverId:%s error(%v)", pushArg, c.serverID, err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// Close close the resources.
func (c *Comet) Close() (err error) {
	finish := make(chan bool)
	go func() {
		for {
			n := len(c.broadcastChan)
			for _, ch := range c.pushChan {
				n += len(ch)
			}
			for _, ch := range c.roomChan {
				n += len(ch)
			}
			if n == 0 {
				finish <- true
				return
			}
			time.Sleep(time.Second)
		}
	}()
	select {
	case <-finish:
		log.Info("close comet finish")
	case <-time.After(5 * time.Second):
		err = fmt.Errorf("close comet(server:%s push:%d room:%d broadcast:%d) timeout", c.serverID, len(c.pushChan), len(c.roomChan), len(c.broadcastChan))
	}
	c.cancel()
	return
}
