package job

import (
	"context"
	"fmt"
	pb "github.com/Terry-Mao/goim/api/logic"
	"github.com/Terry-Mao/goim/internal/job/conf"
	"github.com/golang/protobuf/proto"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"sync"

	cluster "github.com/bsm/sarama-cluster"
	log "github.com/golang/glog"
)

// Job is push job.
type Job struct {
	c            *conf.Config
	consumer     *cluster.Consumer
	cometServers map[string]*Comet

	rooms      map[string]*Room
	roomsMutex sync.RWMutex
}

// New new a push job.
func New(c *conf.Config) *Job {
	j := &Job{
		c:        c,
		consumer: newKafkaSub(c.Kafka),
		rooms:    make(map[string]*Room),
	}
	j.watchComet(c)
	return j
}

func newKafkaSub(c *conf.Kafka) *cluster.Consumer {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	consumer, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Topic}, config)
	if err != nil {
		panic(err)
	}
	return consumer
}

// Close close resounces.
func (j *Job) Close() error {
	if j.consumer != nil {
		return j.consumer.Close()
	}
	return nil
}

// Consume messages, watch signals
func (j *Job) Consume() {
	for {
		select {
		case err := <-j.consumer.Errors():
			log.Errorf("consumer error(%v)", err)
		case n := <-j.consumer.Notifications():
			log.Infof("consumer rebalanced(%v)", n)
		case msg, ok := <-j.consumer.Messages():
			if !ok {
				return
			}
			j.consumer.MarkOffset(msg, "")
			// process push message
			pushMsg := new(pb.PushMsg)
			//pushMsg := &pb.PushMsg{
			//	Type:      pb.PushMsg_PUSH,
			//	Operation: op,     //1000,
			//	Server:    server, //test3
			//	Keys:      keys,   //[7c65e04f-1758-4102-87bb-bcdc9a9b2d0d],
			//	Msg:       msg,    //[]
			//}
			if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
				log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
				continue
			}
			if err := j.push(context.Background(), pushMsg); err != nil {
				log.Errorf("j.push(%v) error(%v)", pushMsg, err)
			}
			log.Infof("consume: %s/%d/%d\t%s\t%+v", msg.Topic, msg.Partition, msg.Offset, msg.Key, pushMsg)
		}
	}
}

func (j *Job) watchComet(c *conf.Config) {
	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  c.ClientConfg,
		ServerConfigs: c.ServerConfig,
	})

	query := vo.SelectInstancesParam{
		ServiceName: "goim.comet",
		GroupName:   "DEFAULT_GROUP",
		HealthyOnly: true,
	}

	instances, err := namingClient.SelectInstances(query)

	if err != nil {
		panic("watchComet init failed")
	}
	j.newAddress(instances)

}

func (j *Job) newAddress(instances []model.Instance) error {
	if len(instances) == 0 {
		return fmt.Errorf("watchComet instance is empty")
	}
	comets := map[string]*Comet{}
	for _, in := range instances {
		if old, ok := j.cometServers[in.Ip]; ok {
			comets[in.Ip] = old
			continue
		}
		c, err := NewComet(&in, j.c.Comet)
		if err != nil {
			log.Errorf("watchComet NewComet(%+v) error(%v)", in, err)
			return err
		}
		comets[in.Ip] = c
		log.Infof("watchComet AddComet grpc:%+v", in)
	}
	for key, old := range j.cometServers {
		if _, ok := comets[key]; !ok {
			old.cancel()
			log.Infof("watchComet DelComet:%s", key)
		}
	}
	j.cometServers = comets
	return nil
}
