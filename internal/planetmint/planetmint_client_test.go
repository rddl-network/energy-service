package planetmint_test

import (
	"log"
	"testing"

	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/planetmint"
)

func TestPlanetmintQueryAccount(t *testing.T) {
	// skipped because test is just to showcase interfaceRegistry fix
	t.SkipNow()

	cfg := config.DefaultConfig()
	grpcConn, err := planetmint.SetupGRPCConnection(cfg)
	if err != nil {
		log.Fatalf("fatal error opening grpc connection %s", err)
	}
	cfg.Planetmint.Actor = "plmnt1p445cz0hfg4yg3dgrq5n3e9wdr8rwpt9qfcz2y"
	_ = planetmint.NewPlanetmintClient(cfg.Planetmint.Actor, grpcConn)
}
