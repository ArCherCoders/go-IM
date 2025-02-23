package comet

import (
	"context"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"math/rand"
	"time"

	"github.com/Terry-Mao/goim/api/logic"
	"github.com/Terry-Mao/goim/internal/comet/conf"
	log "github.com/golang/glog"
	"github.com/zhenjl/cityhash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	minServerHeartbeat = time.Minute * 10
	maxServerHeartbeat = time.Minute * 30
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
	grpcMaxSendMsgSize        = 1 << 24
	grpcMaxCallMsgSize        = 1 << 24
	grpcKeepAliveTime         = time.Second * 10
	grpcKeepAliveTimeout      = time.Second * 3
	grpcBackoffMaxDelay       = time.Second * 3
)

func newLogicClient(c *conf.Config) logic.LogicClient {

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  c.ClientConfg,
		ServerConfigs: c.ServerConfig,
	})

	query := vo.SelectOneHealthInstanceParam{
		ServiceName: "goim.logic",
		GroupName:   "DEFAULT_GROUP",
	}

	instances, err := namingClient.SelectOneHealthyInstance(query)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.RPCServer.IdleTimeout))
	defer cancel()
	IP := instances.Ip
	port := instances.Port
	addr := fmt.Sprintf("%s:%d", IP, port)

	conn, err := grpc.DialContext(ctx, addr, []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithInitialWindowSize(grpcInitialWindowSize),
		grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
		grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                grpcKeepAliveTime,
			Timeout:             grpcKeepAliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"round_robin":{"roundRobin":{}}}]}`),
	}...)
	if err != nil {
		panic(err)
	}
	return logic.NewLogicClient(conn)
}

// Server is comet server.
type Server struct {
	c         *conf.Config
	round     *Round    // accept round store
	buckets   []*Bucket // subkey bucket
	bucketIdx uint32
	serverID  string
	rpcClient logic.LogicClient
}

// NewServer returns a new Server.
func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:         c,
		round:     NewRound(c),
		rpcClient: newLogicClient(c),
	}
	// init bucket
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIdx = uint32(c.Bucket.Size) //32
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}
	//host="192.168.31.77"
	//s.serverID=192.168.31.77
	s.serverID = c.Env.Host // ooo
	go s.onlineproc()
	return s
}

// Buckets return all buckets.
func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

// Bucket get the bucket by subkey.
func (s *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketIdx
	if conf.Conf.Debug {
		log.Infof("%s hit channel bucket index: %d use cityhash", subKey, idx)
	}
	return s.buckets[idx]
}

// RandServerHearbeat rand server heartbeat. 10-//30分钟之内的随机时间
func (s *Server) RandServerHearbeat() time.Duration {
	//fmt.Println("RandServerHearbeat=maxServerHeartbeat-minServerHeartbeat", int64(maxServerHeartbeat-minServerHeartbeat))
	//fmt.Println("RandServerHearbeat=rand.Int63n", rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat)))
	//fmt.Println("RandServerHearbeat=time.Duration", minServerHeartbeat+time.Duration(rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat))))
	return (minServerHeartbeat + time.Duration(rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat))))
}

// Close close the server.
func (s *Server) Close() (err error) {
	return
}

func (s *Server) onlineproc() {
	for {
		var (
			allRoomsCount map[string]int32
			err           error
		)
		roomCount := make(map[string]int32)
		for _, bucket := range s.buckets {
			for roomID, count := range bucket.RoomsCount() {
				roomCount[roomID] += count
			}
		}
		if allRoomsCount, err = s.RenewOnline(context.Background(), s.serverID, roomCount); err != nil {
			time.Sleep(time.Second)
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpRoomsCount(allRoomsCount)
		}
		time.Sleep(time.Second * 10)
	}
}
