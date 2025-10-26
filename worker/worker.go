package worker

import (
	"sync"
	"time"

	"github.com/AdarshN-1324/load_balancer-reverse_proxy/server_conf"
)

func CheckBackend(serverpool *server_conf.Server, timeout int) {

	var wg sync.WaitGroup
	for {
		for i := range serverpool.Backends {
			wg.Add(1)
			go serverpool.Backends[i].CheckActive(&wg, serverpool)
		}
		wg.Wait()
		time.Sleep(time.Second * time.Duration(timeout))
	}

}
