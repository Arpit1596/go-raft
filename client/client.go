package client

import (
	"context"
	_ "leader-election/client/resolver"
	"leader-election/proto"
	"log"

	//"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	//"google.golang.org/protobuf/encoding/protojson"
	//"google.golang.org/protobuf/types/known/emptypb"
)

func GetLeader() string {
	logger := log.New(os.Stdout, "", 0)
	conn, _ := grpc.Dial(
		"tcp:///localhost:50052,localhost:50053,localhost:50054",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	c := proto.NewLeaderServiceClient(conn)

	resp, err := c.GetLeader(context.Background(), &emptypb.Empty{})
	if err != nil {
		logger.Println(err)
		return ""
	}
	return resp.LeaderName
}

/*func main() {
	logger := log.New(os.Stdout, "", 0)
	/*handler := func(w http.ResponseWriter, r *http.Request) {
		conn, _ := grpc.Dial(
			"tcp:///10.0.1.10:50082,10.0.1.11:50082,10.0.1.12:50082",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		c := proto.NewLeaderServiceClient(conn)

		resp, err := c.GetLeader(context.Background(), &emptypb.Empty{})
		if err != nil {
			logger.Println(err)
		}

		jsonData, _ := protojson.Marshal(resp)
		w.Write(jsonData)
	}

	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	http.HandleFunc("/leader", handler)
	http.HandleFunc("/health", healthHandler)

	http.ListenAndServe(":8080", nil)*/

// 	conn, _ := grpc.Dial(
// 		"tcp:///localhost:50052,localhost:50053,localhost:50054",
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 	)

// 	c := proto.NewLeaderServiceClient(conn)

// 	resp, err := c.GetLeader(context.Background(), &emptypb.Empty{})
// 	if err != nil {
// 		logger.Println(err)
// 	}

// 	log.Printf(resp.GetMessage())

// }
