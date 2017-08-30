package db

import (
	"testing"
	"fmt"
	"sync"
	"log"
	"github.com/couchbase/sync_gateway/base"
)



func TestRepairBucket(t *testing.T) {

	base.EnableLogKey("RepairBucket")

	bucket := testBucket()
	log.Printf("installing views")
	installViews(bucket)
	log.Printf("installed views")
	testSyncData := syncData{}
	bucket.Add("foo", 0, map[string]interface{}{"foo": "bar", "_sync": testSyncData})

	repairJobWaitGroup := sync.WaitGroup{}

	repairJob := func(doc *document) (transformedDoc *document, transformed bool, err error) {
		defer repairJobWaitGroup.Done()
		log.Printf("repairJob called back")
		return nil, true, nil
	}
	repairBucket := NewRepairBucket(bucket).
		SetDryRun(true).
		AddRepairJob(repairJob)

	repairJobWaitGroup.Add(1)
	err := repairBucket.RepairBucket()

	assertNoError(t, err, fmt.Sprintf("Unexpected error: %v", err))

	log.Printf("waiting for waitgroup to be finished")
	repairJobWaitGroup.Wait()

}
