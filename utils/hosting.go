package utils

import (
	"context"
	"fmt"
	crypto "github.com/libp2p/go-libp2p-crypto"
	ma "github.com/multiformats/go-multiaddr"
	libp2p "gx/ipfs/QmRBaUEQEeFWywfrZJ64QgsmvcqgLSK3VbvGMR2NM2Edpf/go-libp2p"
	rhost "gx/ipfs/QmRBaUEQEeFWywfrZJ64QgsmvcqgLSK3VbvGMR2NM2Edpf/go-libp2p/p2p/host/routed"
	dht "gx/ipfs/QmXbPygnUKAPMwseE5U3hQA7Thn59GVm7pQrhkFV63umT8/go-libp2p-kad-dht"
	pstore "gx/ipfs/QmZ9zH2FnLcxv1xyzFeUpDUeo55xEhZQHgveZijcxr7TLj/go-libp2p-peerstore"
	ds "gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore"
	dsync "gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore/sync"
	host "gx/ipfs/QmfD51tKgJiTMnW9JEiDiPwsCY4mqUoxkhKhBfyW12spTC/go-libp2p-host"
	"log"
)

// MakeDefaultRoutedHost creates a default LibP2P host with a random peer ID listening on the
// given multiaddress.
func MakeDefaultRoutedHost(listenPort int, priv crypto.PrivKey, useGlobalIPFSInstance bool) (host.Host, error) {
	var bootstrapPeers []pstore.PeerInfo
	if useGlobalIPFSInstance {
		bootstrapPeers = getLocalPeerInfo()
	} else {
		bootstrapPeers = ipfsPeers
	}

	return MakeRoutedHost(listenPort, priv, bootstrapPeers)
}

// MakeRoutedHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true. It will bootstrap using the
// provided PeerInfo
func MakeRoutedHost(listenPort int, priv crypto.PrivKey, bootstrapPeers []pstore.PeerInfo) (host.Host, error) {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
		libp2p.Identity(priv),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	}

	ctx := context.Background()

	basicHost, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// Construct a datastore (needed by the DHT). This is just a simple, in-memory thread-safe datastore.
	dstore := dsync.MutexWrap(ds.NewMapDatastore())

	// Make the DHT
	dht := dht.NewDHT(ctx, basicHost, dstore)

	// Make the routed host
	routedHost := rhost.Wrap(basicHost, dht)

	// connect to the chosen ipfs nodes
	if len(bootstrapPeers) != 0 {
		err = bootstrapConnect(ctx, routedHost, bootstrapPeers)
		if err != nil {
			return nil, err
		}
	}

	// Bootstrap the host
	err = dht.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", routedHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	// addr := routedHost.Addrs()[0]
	addrs := routedHost.Addrs()
	log.Println("I can be reached at:")
	for _, addr := range addrs {
		log.Println(addr.Encapsulate(hostAddr))
	}

	return routedHost, nil
}
