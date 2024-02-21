# go-raft

It uses Hashicorp raft api to elect leader at runtime and whichever is the leader do the work and rest of the instances wait for them to become the leader.

It also has the grpc client which you can call as a customer to get the current leader of the cluster and that grpc client supports passing of multiple ips in a tcp request, at runtime it decides to which ip it wanna resolve.
