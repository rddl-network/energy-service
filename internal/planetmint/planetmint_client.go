package planetmint

import (
	"context"
	"log"

	"github.com/cosmos/cosmos-sdk/codec"
	ctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/planetmint/planetmint-go/lib"
	dertypes "github.com/planetmint/planetmint-go/x/der/types"
	"github.com/rddl-network/energy-service/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type IPlanetmintClient interface {
	RegisterDER(zigbee_id string, plmntAddress string, lidquidAddress string, metadatajson string) error
	IsZigbeeRegistered(zigbee_id string) (bool, error)
}

type PlanetmintClient struct {
	actor string
	conn  *grpc.ClientConn
}

func NewPlanetmintClient(actor string, conn *grpc.ClientConn) *PlanetmintClient {
	return &PlanetmintClient{
		actor: actor,
		conn:  conn,
	}
}

func SetupGRPCConnection(cfg *config.Config) (conn *grpc.ClientConn, err error) {
	interfaceRegistry := ctypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface(
		"cosmos.auth.IAccount",
		(*authtypes.AccountI)(nil),
		&authtypes.BaseAccount{},
		&authtypes.ModuleAccount{},
	)

	return grpc.Dial(
		cfg.Planetmint.RPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(interfaceRegistry).GRPCCodec())),
	)
}

func (pmc *PlanetmintClient) RegisterDER(zigbee_id string, plmntAddress string, lidquidAddress string, metadatajson string) error {
	der := dertypes.DER{
		ZigbeeID:      zigbee_id,
		PlmntAddress:  plmntAddress,
		LiquidAddress: lidquidAddress,
		MetadataJson:  metadatajson,
	}

	// Create the message
	msg := dertypes.NewMsgRegisterDER(pmc.actor, &der)

	// Get the address of the actor
	addr := sdk.MustAccAddressFromBech32(pmc.actor)

	// Broadcast the transaction
	if _, err := lib.BroadcastTxWithFileLock(addr, msg); err != nil {
		return err
	}
	log.Printf("[DEBUG] RegisterDER: Successfully registered DER for ZigbeeID %s", zigbee_id)
	return nil
}

func (pmc *PlanetmintClient) IsZigbeeRegistered(zigbeeID string) (registered bool, err error) {
	derClient := dertypes.NewQueryClient(pmc.conn)
	res, err := derClient.Der(context.Background(), &dertypes.QueryDerRequest{ZigbeeID: zigbeeID})
	if err != nil {
		return
	}
	if res != nil && res.Der != nil {
		registered = res.Der.ZigbeeID == zigbeeID
		if registered {
			log.Printf("[DEBUG] IsZigbeeRegistered: ZigbeeID %s is registered", zigbeeID)
		} else {
			log.Printf("[DEBUG] IsZigbeeRegistered: ZigbeeID %s is not registered", zigbeeID)
		}
	} else {
		log.Printf("[DEBUG] IsZigbeeRegistered: No DER found for ZigbeeID %s", zigbeeID)
	}
	return
}
