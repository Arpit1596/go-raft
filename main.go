package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"leader-election/proto"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	leader "leader-election/leader"

	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	myAddr           = "localhost" + ":" + os.Getenv("PORT")
	raftId           = os.Getenv("SERVER_ID")
	raftDir          = "/tmp/my-raft-cluster/3"
	raftBootstrap, _ = strconv.ParseBool(os.Getenv("BOOTSTRAP_CLUSTER"))
	raftLeader       = os.Getenv("LEADER_ADDRESS")
)

var wasLeader = false

type raftServer struct {
	r *raft.Raft
}

func (s *raftServer) AddNode(ctx context.Context, req *proto.Command) (*proto.CommandResponse, error) {
	s.r.AddVoter(raft.ServerID(req.ServerId), raft.ServerAddress(req.ServerAddress), 0, 0)
	response := &proto.CommandResponse{
		Result: fmt.Sprintf("Added node %s to the cluster", req.ServerId),
	}

	return response, nil
}

type leaderServer struct {
	l *leader.Leader
	r *raft.Raft
}

func (s *leaderServer) GetLeader(ctx context.Context, _ *emptypb.Empty) (*proto.Leader, error) {
	s.l.Mtx.RLock()
	defer s.l.Mtx.RUnlock()
	_, serverId := s.r.LeaderWithID()
	if len(string(serverId)) == 0 {
		return &proto.Leader{LeaderName: "No leader, I am the only server up."}, nil
	} else {
		return &proto.Leader{LeaderName: string(serverId)}, nil
	}

}

func main() {
	flag.Parse()

	if raftId == "" {
		log.Fatalf("flag --raft_id is required")
	}

	// if !raftBootstrap && raftLeader == "" {
	// 	log.Fatalf("flag --raft_leader is required when adding a peer")
	// }
	// if !raftBootstrap {
	// 	leader := client.GetLeader()
	// 	if len(leader) == 0 && raftLeader == ""{
	// 		log.Fatalf("flag --raft_leader is required when adding a peer")
	// 	}else{
	// 		raftLeader = leader
	// 	}
	// }

	isEmpty, err := isDirEmpty(raftDir)
	if err != nil {
		log.Fatalf("failed to check raft dir", err)
	}
	println("heeheeeeeeeeeee", isEmpty)
	if !isEmpty {
		raftBootstrap = false
	}

	host, port, err := net.SplitHostPort(myAddr)
	if err != nil {
		log.Fatalf("failed to parse local address (%q): %v", port, err)
	}

	portNum, _ := strconv.Atoi(port)
	sock, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, strconv.Itoa(portNum+1)))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	leaderStruct := &leader.Leader{}
	r, err := NewRaft(raftId, myAddr, leaderStruct)
	if err != nil {
		log.Fatalf("failed to start raft: %v", err)
	}

	Report(r)

	s := grpc.NewServer()
	proto.RegisterCommandServiceServer(s, &raftServer{r})
	proto.RegisterLeaderServiceServer(s, &leaderServer{l: leaderStruct, r: r})
	if err := s.Serve(sock); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func NewRaft(myID, myAddress string, fsm raft.FSM) (*raft.Raft, error) {
	c := raft.DefaultConfig()
	c.LocalID = raft.ServerID(myID)

	baseDir := raftDir
	ldb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "logs.db"))
	if err != nil {
		return nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "logs.db"), err)
	}

	sdb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "stable.db"))
	if err != nil {
		return nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "stable.db"), err)
	}

	fss, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, baseDir, err)
	}

	tm, err := raft.NewTCPTransport(myAddress, nil, 3, 10*time.Second, os.Stderr)

	if err != nil {
		return nil, fmt.Errorf("raft.NewTCPTransport: %v", err)
	}

	r, err := raft.NewRaft(c, fsm, ldb, sdb, fss, tm)
	if err != nil {
		return nil, fmt.Errorf("raft.NewRaft: %v", err)
	}

	_, id := r.LeaderWithID()
	println("Leader before", id)

	if raftBootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(myID),
					Address:  raft.ServerAddress(myID),
				},
			},
		}
		f := r.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			return nil, fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
		}
	} else {
		host, port, _ := net.SplitHostPort(raftLeader)
		portNum, _ := strconv.Atoi(port)
		listener, _ := grpc.Dial(
			fmt.Sprintf("%s:%s", host, strconv.Itoa(portNum+1)),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		request := &proto.Command{
			Command:       "AddVoter",
			ServerId:      myID,
			ServerAddress: myID,
			PrevIndex:     0,
		}
		c := proto.NewCommandServiceClient(listener)
		c.AddNode(context.Background(), request)

	}

	return r, nil
}

func Report(r *raft.Raft) {
	ch := make(chan raft.Observation, 1)
	r.RegisterObserver(raft.NewObserver(ch, true, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.LeaderObservation)
		return ok
	}))
	setServingStatus(r, r.State() == raft.Leader)
	go func() {
		for data := range ch {
			if len(data.Data.(raft.LeaderObservation).LeaderID) >0 {
				log.Println("Leader is changed hence ", r.State() == raft.Leader)
				_, id := r.LeaderWithID()
				log.Println("Leader after", id)
				setServingStatus(r, r.State() == raft.Leader)
			}else{
				
				log.Println("no leader forund", r.String())
			}
			
			log.Println("=================", r.Stats())
		}
	}()

	
}

func removeServer(r *raft.Raft, isLeader bool) {
	if isLeader {
		wasLeader = true
		ch := make(chan raft.Observation, 1)
		r.RegisterObserver(raft.NewObserver(ch, true, func(o *raft.Observation) bool {
			_, ok := o.Data.(raft.FailedHeartbeatObservation)
			return ok
		}))

		go func() {
			for value := range ch {
				log.Println("Failed Node", value.Data.(raft.FailedHeartbeatObservation).PeerID)
				r.RemoveServer(value.Data.(raft.FailedHeartbeatObservation).PeerID, 0, 0)
			}
		}()
	} else {
		if wasLeader {
			os.Exit(1)
		}
	}
}

func setServingStatus(r *raft.Raft, isLeader bool) {
	log.Println("Leader is changed hence ", isLeader)
	if isLeader {
		r.ApplyLog(raft.Log{Data: []byte(raftId)}, time.Minute)
		go leaderWork()
	}
}

func leaderWork() {
	for {
		log.Println("I am working!")
		time.Sleep(10 * time.Second)
	}
}

func getContainerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error getting local IP addresses:", err)
		return ""
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			// Display the local IPv4 address of the host
			fmt.Println("Local IP address of the host:", ipNet.IP.String())
			return ipNet.IP.String()
		}
	}
	return ""
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
