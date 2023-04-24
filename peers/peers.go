package peers

import "github.com/marsxingzhi/marscache/pb"

// 根据key获取对应的节点(节点能力)
type PeerPicker interface {
	PickPeer(key string) (getter PeerGetter, ok bool)
}

// 从group中获取key对应的值
type PeerGetter interface {
	// Get(group string, key string) ([]byte, error)
	Get(in *pb.Request, out *pb.Response) error
}
