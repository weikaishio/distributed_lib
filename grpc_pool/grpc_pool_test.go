package grpc_pool

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"
	"reflect"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/credentials"
	"github.com/weikaishio/distributed_lib/etcd"
	)

const (
	HOST = "127.0.0.1"
	PORT = 8890
	CER_FILE      = ""
	KEY_FILE      = ""
	RPC_AUTH_NAME = "test"
	RPC_AUTH_PWD  = "test"
	SERVICE_NAME = ""
	SERVER_Name=""
)

var (
	discovery *etcd.Discovery
	rpcPool *RPCPool
	etcdEndPoints                           []string
	etcdUsername, etcdPassword, serviceName string
)

func init() {
	go svrRun()
	//1、use discovery
	//var err error
	//discovery, err = etcd.NewDiscovery(etcdEndPoints, etcdUsername, etcdPassword)
	//if err != nil {
	//	log.Fatal(fmt.Sprintf("etcd.NewDiscovery etcdEndPoints:%v,err:%v", etcdEndPoints, err))
	//
	//}
	//discovery.Register(serviceName, HOST, PORT, 60*time.Second, "gRPC")
	//discovery.UnRegister(serviceName, HOST, PORT)
	time.Sleep(10 * time.Microsecond)
}
func makeCliPool(){
	//1、use discovery
	//makeConn := func() (*grpc.ClientConn, error) {
	//	return discovery.DialWithAuth(SERVICE_NAME, RPC_AUTH_NAME, RPC_AUTH_PWD, SERVER_Name, CER_FILE)
	//}
	//2、normal dial addr
	makeConn := func() (Connection, error) {
		//return grpc.Dial(fmt.Sprintf("127.0.0.1:%d", PORT), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(1000)*time.Millisecond))
		return etcd.DialWithAuthWithAddr(fmt.Sprintf("127.0.0.1:%d", PORT),RPC_AUTH_NAME, RPC_AUTH_PWD,SERVER_Name, CER_FILE)
	}
	rpcPool = NewRPCPool(makeConn)
	rpcPool.InitPool()
}

func svrRun() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatal(fmt.Sprintf("RPCSvrRun net.Listen err:%v", err))
	}
	var opts []grpc.ServerOption
	if CER_FILE!="" && KEY_FILE!="" {
		creds, err := credentials.NewServerTLSFromFile(CER_FILE, KEY_FILE)
		if err != nil {
			log.Fatal(fmt.Sprintf("RPCSvrRun Failed to generate credentials %v", err))
		}
		opts=append(opts,grpc.Creds(creds))
	}
	opts=append(opts,grpc.StreamInterceptor(streamInterceptor))
	opts=append(opts,grpc.UnaryInterceptor(grpc.UnaryServerInterceptor(unaryInterceptor)))
	rpcSvr := grpc.NewServer(
		opts...
	)
	RegisterGrpcTestServer(rpcSvr, GrpcTestServer(&TestServer{}))
	fmt.Println("gRPC Run")
	if err = rpcSvr.Serve(listen); err != nil {
		log.Fatal(fmt.Sprintf("RPCSvrRun svr.Serve err:%v", err))
	}
}
func streamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := authorize(stream.Context()); err != nil {
		return err
	}
	return handler(srv, stream)
}

func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := authorize(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func authorize(ctx context.Context) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md["username"]) > 0 && md["username"][0] == RPC_AUTH_NAME &&
			len(md["password"]) > 0 && md["password"][0] == RPC_AUTH_PWD {
			return nil
		}

		return errors.New("auth fail")
	}

	return errors.New("auth empty")
}

type TestServer struct {
}

func (t *TestServer) HelloWorld(ctx context.Context, in *HelloWorldModel) (*CommonResp, error) {
	return &CommonResp{
		Status: ReturnStatus_StatusSuccess,
		Desc:   in.Title,
	}, nil
}

func TestRPCPool_Borrow(t *testing.T) {
	makeCliPool()

	conn := rpcPool.Borrow()
	if conn == nil {
		t.Log("执行出错，无可用连接")
		return
	} else {
		defer rpcPool.Return(conn)
	}

	cli := NewGrpcTestClient(conn.(*grpc.ClientConn))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//1、normal invoke
	//resp, err := cli.HelloWorld(ctx, &HelloWorldModel{Title: "Hello World"})
	//t.Logf("resp.Status:%v,resp.Desc:%v,err:%v", resp.Status, resp.Desc, err)

	//2、use reflect
	reflectCli := reflect.ValueOf(cli)
	reflectMethod := reflectCli.MethodByName("HelloWorld")
	if !reflectMethod.IsValid() {
		t.Logf("!reflectMethod.IsValid()")
		return
	}

	in:=&HelloWorldModel{Title: "Hello World"}
	vals := reflectMethod.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(in)})

	var (
		err  error
		resp *CommonResp
	)
	if !vals[0].IsNil() {
		resp = vals[0].Interface().(*CommonResp)
		t.Logf("resp.Status:%v,resp.Desc:%v", resp.Status, resp.Desc)
	}
	if !vals[1].IsNil() {
		err = vals[1].Interface().(error)
		t.Logf("err:%v", err)
	}
}
