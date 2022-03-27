package dal

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/monitor/models"
	"git.d464.sh/adc/telemetry/telemetry/snapshot"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// upserts the given session and returns its ID
func Session(ctx context.Context, exec boil.ContextExecutor, p peer.ID, s uuid.UUID) (int, error) {
	sess := models.Session{
		SessionUUID: s.String(),
		Peerid:      p.String(),
		LastSeen:    time.Now(),
	}
	err := sess.Upsert(ctx, exec, true,
		[]string{models.SessionColumns.SessionUUID, models.SessionColumns.Peerid},
		boil.Whitelist(models.SessionColumns.LastSeen),
		boil.Infer(),
	)
	if err != nil {
		return 0, err
	}
	return sess.SessionID, nil
}

func RoutingTable(ctx context.Context, exec boil.ContextExecutor, sessionID int, snapshot *snapshot.Snapshot) error {
	ssrt := snapshot.GetRoutingTable()
	buckets := make([]int64, 0, len(ssrt.Buckets))
	for _, b := range ssrt.Buckets {
		buckets = append(buckets, int64(b))
	}

	rt := models.SnapshotsRT{
		SessionID:    sessionID,
		SnapshotTime: snapshot.Time,
		Buckets:      buckets,
	}
	return rt.Insert(ctx, exec, boil.Infer())
}
