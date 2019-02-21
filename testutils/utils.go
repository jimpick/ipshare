package testutils

import (
	"context"
	"log"

	peerstore "gx/ipfs/QmZ9zH2FnLcxv1xyzFeUpDUeo55xEhZQHgveZijcxr7TLj/go-libp2p-peerstore"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	host "gx/ipfs/QmfD51tKgJiTMnW9JEiDiPwsCY4mqUoxkhKhBfyW12spTC/go-libp2p-host"

	mrand "math/rand"

	multihash "github.com/multiformats/go-multihash"
	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"

	lutils "github.com/aschmahmann/ipshare/utils"
)

func createHost(rnd *mrand.Rand, portNum int) (host.Host, error) {
	priv, _, err := crypto.GenerateEd25519Key(rnd)
	if err != nil {
		return nil, err
	}

	const localDaemon = true

	ha, err := lutils.MakeRoutedHost(portNum, priv, nil)
	if err != nil {
		return nil, err
	}

	return ha, nil
}

func CreateHostAndPeers(rnd *mrand.Rand, startPort, numHosts int, printPeers bool) ([]host.Host, []peer.ID, error) {
	var hosts []host.Host
	var peers []peer.ID

	hBase, err := createHost(rnd, 0)
	if err != nil {
		return nil, nil, err
	}
	hBasePeerInfo := peerstore.PeerInfo{Addrs: hBase.Addrs(), ID: hBase.ID()}

	for i := 0; i < numHosts; i++ {
		h, err := createHost(rnd, 0)
		if err != nil {
			return nil, nil, err
		}

		err = h.Connect(context.Background(), hBasePeerInfo)
		if err != nil {
			return nil, nil, err
		}

		hosts = append(hosts, h)
		peers = append(peers, h.ID())
	}

	if printPeers {
		for i, p := range peers {
			log.Printf("peer%v: %v", i, p)
		}
	}

	return hosts, peers, nil
}

var cidBuilder = cid.V1Builder{Codec: cid.Raw, MhType: multihash.SHA2_256, MhLength: -1}

func CreateCid(data string) cid.Cid {
	c, err := cidBuilder.Sum([]byte(data))
	if err != nil {
		panic(err)
	}
	return c
}
