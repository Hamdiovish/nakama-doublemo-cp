package server

import (
	"context"
	"encoding/binary"
	"math/rand"
	"strconv"
	"sync"
	"time"

	nakamacluster "github.com/doublemo/nakama-cluster"
	ncapi "github.com/doublemo/nakama-cluster/api"
	"github.com/doublemo/nakama-cluster/endpoint"
	etcdv3 "github.com/doublemo/nakama-cluster/sd/etcdv3"
	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ClusterServer struct {
	ctx             context.Context
	cancelFn        context.CancelFunc
	client          *nakamacluster.Client
	config          Config
	tracker         Tracker
	sessionRegistry SessionRegistry
	statusRegistry  *StatusRegistry
	partyRegistry   PartyRegistry
	matchRegistry   MatchRegistry
	logger          *zap.Logger
	once            sync.Once
}

var cc *ClusterServer

func CC() *ClusterServer {
	return cc
}

func StartClusterServer(ctx context.Context, logger *zap.Logger, config Config) *ClusterServer {
	logger.Info("Initializing cluster")
	clusterConfig := config.GetCluster()
	options := etcdv3.ClientOptions{
		Cert:          clusterConfig.Etcd.Cert,
		Key:           clusterConfig.Etcd.Key,
		CACert:        clusterConfig.Etcd.CACert,
		DialTimeout:   time.Duration(clusterConfig.Etcd.DialTimeout),
		DialKeepAlive: time.Duration(clusterConfig.Etcd.DialKeepAlive),
		Username:      clusterConfig.Etcd.Username,
		Password:      clusterConfig.Etcd.Password,
	}

	sdClient, err := etcdv3.NewClient(context.Background(), clusterConfig.Etcd.Endpoints, options)
	if err != nil {
		logger.Fatal("Failed initializing etcd", zap.Error(err))
	}

	vars := map[string]string{
		"rpc_port": strconv.Itoa(config.GetSocket().Port - 1),
		"port":     strconv.Itoa(config.GetSocket().Port),
	}

	ctx, cancelFn := context.WithCancel(ctx)
	s := &ClusterServer{
		ctx:      ctx,
		cancelFn: cancelFn,
		config:   config,
		logger:   logger,
	}

	client := nakamacluster.NewClient(ctx, logger, sdClient, config.GetName(), vars, clusterConfig.Config)
	client.OnDelegate(s)
	s.client = client
	cc = s
	return s
}

func (s *ClusterServer) NotifyAlive(node endpoint.Endpoint) error {
	return nil
}

func (s *ClusterServer) LocalState(join bool) []byte {
	return nil
}

func (s *ClusterServer) MergeRemoteState(buf []byte, join bool) {}

func (s *ClusterServer) NotifyJoin(node endpoint.Endpoint) {}

func (s *ClusterServer) NotifyLeave(node endpoint.Endpoint) {}

func (s *ClusterServer) NotifyUpdate(node endpoint.Endpoint) {}

func (s *ClusterServer) NotifyMsg(node string, msg []byte) []byte {
	// Check the payload type using a switch statement
	//switch msg.Payload.(type) {
	// case *ncapi.Envelope_Message:
	// 	s.onMessage(node, msg)

	// case *ncapi.Envelope_SessionNew:
	// 	s.onSessionUp(node, msg)

	// case *ncapi.Envelope_SessionClose:
	// 	s.onSessionDown(node, msg)

	//case *ncapi.Message_Envelope:
	//	s.onTrack(node, msg.)

	var message ncapi.Message_Track
	if err := proto.Unmarshal(msg, &message); err != nil {
		s.logger.Warn("NotifyMsg parse failed", zap.Error(err))
	} else {
		s.logger.Error("message: ", zap.Any("Payload", message.Presences))
		s.onTrack(node, message)
	}

	// var untrack_message ncapi.Message_Untrack
	// if err := proto.Unmarshal(msg, &untrack_message); err != nil {
	// 	s.logger.Warn("NotifyMsg parse failed", zap.Error(err))
	// } else {
	// 	s.onUntrack(node, untrack_message)
	// }

	// case *ncapi.Envelope_Untrack:
	// 	s.onUntrack(node, msg)

	// case *ncapi.Envelope_UntrackAll:
	// 	s.onUntrackAll(node, msg)

	// case *ncapi.Envelope_UntrackByMode:
	// 	s.onUntrackByMode(node, msg)

	// case *ncapi.Envelope_UntrackByStream:
	// 	s.onUntrackByStream(node, msg)

	// case *ncapi.Envelope_Bytes:
	// 	return s.onBytes(node, msg)

	//}
	return nil
}

// Send 使用TCP发送信息
func (s *ClusterServer) SendAndRecv(msg *ncapi.Envelope, to ...string) ([][]byte, error) {
	data := make([]byte, 32)
	binary.BigEndian.PutUint32(data, rand.Uint32())
	return s.client.Send(nakamacluster.NewMessage(msg.GetBytes(), to...))
}

func (s *ClusterServer) Send(msg *ncapi.Envelope, to ...string) error {
	_, err := s.client.Send(nakamacluster.NewMessage(msg.GetBytes()), to...)
	return err
}

func (s *ClusterServer) Broadcast(msg *ncapi.Envelope) error {
	return s.client.Broadcast(nakamacluster.NewMessage(msg.GetBytes()))
}

// func (s *ClusterServer) RPCCall(ctx context.Context, name, key, cid string, vars map[string]string, in []byte) ([]byte, error) {
// 	return s.client.RPCCall(ctx, name, key, cid, vars, in)
// }

func (s *ClusterServer) NodeId() string {
	return s.client.GetLocalNode().Name
}

// func (s *ClusterServer) NodeStatus() nakamacluster.MetaStatus {
// 	return s.client.GetMeta().Status
// }

func (s *ClusterServer) Stop() {
	s.once.Do(func() {
		if s.cancelFn != nil {
			s.client.Stop()
			s.cancelFn()
		}
	})
}

func (s *ClusterServer) SetTracker(t Tracker) {
	s.tracker = t
}

func (s *ClusterServer) SetSessionRegistry(sessionRegistry SessionRegistry) {
	s.sessionRegistry = sessionRegistry
}

func (s *ClusterServer) SetStatusRegistry(statusRegistry *StatusRegistry) {
	s.statusRegistry = statusRegistry
}

func (s *ClusterServer) SetPartyRegistry(partyRegistry PartyRegistry) {
	s.partyRegistry = partyRegistry
}

func (s *ClusterServer) SetMatchRegistry(matchRegistry MatchRegistry) {
	s.matchRegistry = matchRegistry
}

// Method to get nodes from the nodes map
func (s *ClusterServer) GetNodes() []*memberlist.Node {

	// Create a slice to hold the nodes
	nodes := make([]*memberlist.Node, 0, len(s.client.GetNodes()))

	// Iterate over the map and collect the nodes
	for _, node := range nodes {
		nodes = append(nodes, node)
	}

	return nodes
}
