package db

import (
	"os"
	"testing"

	"github.com/couchbase/sync_gateway/base"
)

func TestMain(m *testing.M) {
	base.GTestBucketPool = base.NewTestBucketPool(ViewsAndGSIBucketReadier, ViewsAndGSIBucketInit)

	status := m.Run()

	base.GTestBucketPool.Close()

	os.Exit(status)
}
