package db

import (
	"fmt"
	"testing"

	sgbucket "github.com/couchbase/sg-bucket"
	"github.com/couchbase/sync_gateway/base"
	"github.com/couchbase/sync_gateway/channels"
	goassert "github.com/couchbaselabs/go.assert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Validate stats for view query
func TestQueryChannelsStatsView(t *testing.T) {

	if !base.UnitTestUrlIsWalrus() {
		t.Skip("This test is walrus-only (requires views)")
	}

	db, testBucket := setupTestDB(t)
	defer testBucket.Close()
	defer tearDownTestDB(t, db)

	// docID -> Sequence
	docSeqMap := make(map[string]uint64, 3)

	_, doc, err := db.Put("queryTestDoc1", Body{"channels": []string{"ABC"}})
	require.NoError(t, err, "Put queryDoc1")
	docSeqMap["queryTestDoc1"] = doc.Sequence
	_, doc, err = db.Put("queryTestDoc2", Body{"channels": []string{"ABC"}})
	require.NoError(t, err, "Put queryDoc2")
	docSeqMap["queryTestDoc2"] = doc.Sequence
	_, doc, err = db.Put("queryTestDoc3", Body{"channels": []string{"ABC"}})
	require.NoError(t, err, "Put queryDoc3")
	docSeqMap["queryTestDoc3"] = doc.Sequence

	// Check expvar prior to test
	queryCountExpvar := fmt.Sprintf(base.StatKeyViewQueryCountExpvarFormat, DesignDocSyncGateway(), ViewChannels)
	queryTimeExpvar := fmt.Sprintf(base.StatKeyViewQueryTimeExpvarFormat, DesignDocSyncGateway(), ViewChannels)
	errorCountExpvar := fmt.Sprintf(base.StatKeyViewQueryErrorCountExpvarFormat, DesignDocSyncGateway(), ViewChannels)

	channelQueryCountBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryCountExpvar))
	channelQueryTimeBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryTimeExpvar))
	channelQueryErrorCountBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(errorCountExpvar))

	// Issue channels query
	results, queryErr := db.QueryChannels("ABC", docSeqMap["queryTestDoc1"], docSeqMap["queryTestDoc3"], 100, false)
	assert.NoError(t, queryErr, "Query error")

	assert.Equal(t, 3, countQueryResults(results))

	closeErr := results.Close()
	assert.NoError(t, closeErr, "Close error")

	channelQueryCountAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryCountExpvar))
	channelQueryTimeAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryTimeExpvar))
	channelQueryErrorCountAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(errorCountExpvar))

	assert.Equal(t, channelQueryCountBefore+1, channelQueryCountAfter)
	assert.True(t, channelQueryTimeAfter > channelQueryTimeBefore, "Channel query time stat didn't change")
	assert.Equal(t, channelQueryErrorCountBefore, channelQueryErrorCountAfter)

}

// Validate stats for n1ql query
func TestQueryChannelsStatsN1ql(t *testing.T) {

	if base.UnitTestUrlIsWalrus() {
		t.Skip("This test is Couchbase Server only")
	}

	db, testBucket := setupTestDB(t)
	defer testBucket.Close()
	defer tearDownTestDB(t, db)

	// docID -> Sequence
	docSeqMap := make(map[string]uint64, 3)

	_, doc, err := db.Put("queryTestDoc1", Body{"channels": []string{"ABC"}})
	require.NoError(t, err, "Put queryDoc1")
	docSeqMap["queryTestDoc1"] = doc.Sequence
	_, doc, err = db.Put("queryTestDoc2", Body{"channels": []string{"ABC"}})
	require.NoError(t, err, "Put queryDoc2")
	docSeqMap["queryTestDoc2"] = doc.Sequence
	_, doc, err = db.Put("queryTestDoc3", Body{"channels": []string{"ABC"}})
	require.NoError(t, err, "Put queryDoc3")
	docSeqMap["queryTestDoc3"] = doc.Sequence

	// Check expvar prior to test
	queryCountExpvar := fmt.Sprintf(base.StatKeyN1qlQueryCountExpvarFormat, QueryTypeChannels)
	queryTimeExpvar := fmt.Sprintf(base.StatKeyN1qlQueryTimeExpvarFormat, QueryTypeChannels)
	errorCountExpvar := fmt.Sprintf(base.StatKeyN1qlQueryErrorCountExpvarFormat, QueryTypeChannels)

	channelQueryCountBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryCountExpvar))
	channelQueryTimeBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryTimeExpvar))
	channelQueryErrorCountBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(errorCountExpvar))

	// Issue channels query
	results, queryErr := db.QueryChannels("ABC", docSeqMap["queryTestDoc1"], docSeqMap["queryTestDoc3"], 100, false)
	assert.NoError(t, queryErr, "Query error")

	assert.Equal(t, 3, countQueryResults(results))

	closeErr := results.Close()
	assert.NoError(t, closeErr, "Close error")

	channelQueryCountAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryCountExpvar))
	channelQueryTimeAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryTimeExpvar))
	channelQueryErrorCountAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(errorCountExpvar))

	assert.Equal(t, channelQueryCountBefore+1, channelQueryCountAfter)
	assert.True(t, channelQueryTimeAfter > channelQueryTimeBefore, "Channel query time stat didn't change")
	assert.Equal(t, channelQueryErrorCountBefore, channelQueryErrorCountAfter)

}

// Validate query and stats for sequence view query
func TestQuerySequencesStatsView(t *testing.T) {

	db, testBucket := setupTestDBWithViewsEnabled(t)
	defer testBucket.Close()
	defer tearDownTestDB(t, db)

	// docID -> Sequence
	docSeqMap := make(map[string]uint64, 20)

	// Add docs without channel assignment (will only be assigned to the star channel)
	for i := 1; i <= 10; i++ {
		docID := fmt.Sprintf("queryTestDoc%d", i)
		_, doc, err := db.Put(docID, Body{"nochannels": true})
		assert.NoError(t, err, "Put queryDoc")
		docSeqMap[docID] = doc.Sequence
	}

	// Check expvar prior to test
	queryCountExpvar := fmt.Sprintf(base.StatKeyViewQueryCountExpvarFormat, DesignDocSyncGateway(), ViewChannels)
	errorCountExpvar := fmt.Sprintf(base.StatKeyViewQueryErrorCountExpvarFormat, DesignDocSyncGateway(), ViewChannels)

	channelQueryCountBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryCountExpvar))
	channelQueryErrorCountBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(errorCountExpvar))

	// Issue channels query
	results, queryErr := db.QuerySequences([]uint64{
		docSeqMap["queryTestDoc3"], docSeqMap["queryTestDoc4"],
		docSeqMap["queryTestDoc6"], docSeqMap["queryTestDoc8"],
	})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 4)
	closeErr := results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with single key
	results, queryErr = db.QuerySequences([]uint64{docSeqMap["queryTestDoc2"]})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 1)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with key outside keyset range
	results, queryErr = db.QuerySequences([]uint64{100})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 0)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with empty keys
	results, queryErr = db.QuerySequences([]uint64{})
	assert.Error(t, queryErr, "Expect empty sequence error")

	channelQueryCountAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryCountExpvar))
	channelQueryErrorCountAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(errorCountExpvar))

	goassert.Equals(t, channelQueryCountBefore+3, channelQueryCountAfter)
	goassert.Equals(t, channelQueryErrorCountBefore, channelQueryErrorCountAfter)

	// Add some docs in different channels, to validate query handling when non-star channel docs are present
	for i := 1; i <= 10; i++ {
		docID := fmt.Sprintf("queryTestDocChanneled%d", i)
		_, doc, err := db.Put(docID, Body{"channels": []string{fmt.Sprintf("ABC%d", i)}})
		require.NoError(t, err, "Put queryDoc")
		docSeqMap[docID] = doc.Sequence
	}
	// Issue channels query
	results, queryErr = db.QuerySequences([]uint64{
		docSeqMap["queryTestDoc3"], docSeqMap["queryTestDoc4"],
		docSeqMap["queryTestDoc6"], docSeqMap["queryTestDoc8"],
		docSeqMap["queryTestDocChanneled5"],
	})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 5)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with single key
	results, queryErr = db.QuerySequences([]uint64{docSeqMap["queryTestDoc2"]})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 1)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with key outside sequence range.  Note that this isn't outside the entire view key range, as
	// [*, 25] is sorted before ["ABC1", 11]
	results, queryErr = db.QuerySequences([]uint64{100})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 0)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")
}

// Validate query and stats for sequence view query
func TestQuerySequencesStatsN1ql(t *testing.T) {

	if base.UnitTestUrlIsWalrus() {
		t.Skip("This test is Couchbase Server only")
	}

	db, testBucket := setupTestDB(t)
	defer testBucket.Close()
	defer tearDownTestDB(t, db)

	// docID -> Sequence
	docSeqMap := make(map[string]uint64, 20)

	// Add docs without channel assignment (will only be assigned to the star channel)
	for i := 1; i <= 10; i++ {
		docID := fmt.Sprintf("queryTestDoc%d", i)
		_, doc, err := db.Put(docID, Body{"nochannels": true})
		require.NoError(t, err, "Put queryDoc")
		docSeqMap[docID] = doc.Sequence
	}

	// Check expvar prior to test
	queryCountExpvar := fmt.Sprintf(base.StatKeyN1qlQueryCountExpvarFormat, QueryTypeSequences)
	errorCountExpvar := fmt.Sprintf(base.StatKeyN1qlQueryErrorCountExpvarFormat, QueryTypeSequences)

	channelQueryCountBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryCountExpvar))
	channelQueryErrorCountBefore := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(errorCountExpvar))

	// Issue channels query
	results, queryErr := db.QuerySequences([]uint64{
		docSeqMap["queryTestDoc3"], docSeqMap["queryTestDoc4"],
		docSeqMap["queryTestDoc6"], docSeqMap["queryTestDoc8"],
	})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 4)
	closeErr := results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with single key
	results, queryErr = db.QuerySequences([]uint64{docSeqMap["queryTestDoc2"]})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 1)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with key outside keyset range
	results, queryErr = db.QuerySequences([]uint64{100})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 0)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with empty keys
	results, queryErr = db.QuerySequences([]uint64{})
	assert.Error(t, queryErr, "Expect empty sequence error")

	channelQueryCountAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(queryCountExpvar))
	channelQueryErrorCountAfter := base.ExpvarVar2Int(db.DbStats.StatsGsiViews().Get(errorCountExpvar))

	goassert.Equals(t, channelQueryCountBefore+3, channelQueryCountAfter)
	goassert.Equals(t, channelQueryErrorCountBefore, channelQueryErrorCountAfter)

	// Add some docs in different channels, to validate query handling when non-star channel docs are present
	for i := 1; i <= 10; i++ {
		docID := fmt.Sprintf("queryTestDocChanneled%d", i)
		_, doc, err := db.Put(docID, Body{"channels": []string{fmt.Sprintf("ABC%d", i)}})
		require.NoError(t, err, "Put queryDoc")
		docSeqMap[docID] = doc.Sequence
	}

	// Issue channels query
	results, queryErr = db.QuerySequences([]uint64{
		docSeqMap["queryTestDoc3"], docSeqMap["queryTestDoc4"],
		docSeqMap["queryTestDoc6"], docSeqMap["queryTestDoc8"],
		docSeqMap["queryTestDocChanneled5"],
	})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 5)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with single key
	results, queryErr = db.QuerySequences([]uint64{docSeqMap["queryTestDoc2"]})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 1)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")

	// Issue query with key outside sequence range.  Note that this isn't outside the entire view key range, as
	// [*, 25] is sorted before ["ABC1", 11]
	results, queryErr = db.QuerySequences([]uint64{100})
	assert.NoError(t, queryErr, "Query error")
	goassert.Equals(t, countQueryResults(results), 0)
	closeErr = results.Close()
	assert.NoError(t, closeErr, "Close error")
}

// Validate that channels queries (channels, starChannel) are covering
func TestCoveringQueries(t *testing.T) {
	if base.UnitTestUrlIsWalrus() {
		t.Skip("This test is Couchbase Server only")
	}

	db, testBucket := setupTestDB(t)
	defer testBucket.Close()
	defer tearDownTestDB(t, db)

	gocbBucket, ok := base.AsGoCBBucket(testBucket)
	if !ok {
		t.Errorf("Unable to get gocbBucket for testBucket")
	}

	// channels
	channelsStatement, params := db.buildChannelsQuery("ABC", 0, 10, 100, false)
	plan, explainErr := gocbBucket.ExplainQuery(channelsStatement, params)
	assert.NoError(t, explainErr, "Error generating explain for channels query")
	covered := isCovered(plan)
	planJSON, err := base.JSONMarshal(plan)
	assert.NoError(t, err)
	assert.True(t, covered, "Channel query isn't covered by index: %s", planJSON)

	// star channel
	channelStarStatement, params := db.buildChannelsQuery("*", 0, 10, 100, false)
	plan, explainErr = gocbBucket.ExplainQuery(channelStarStatement, params)
	assert.NoError(t, explainErr, "Error generating explain for star channel query")
	covered = isCovered(plan)
	planJSON, err = base.JSONMarshal(plan)
	assert.NoError(t, err)
	assert.True(t, covered, "Star channel query isn't covered by index: %s", planJSON)

	// Access and roleAccess currently aren't covering, because of the need to target the user property by name
	// in the SELECT.
	// Including here for ease-of-conversion when we get an indexing enhancement to support covered queries.
	accessStatement := db.buildAccessQuery("user1")
	plan, explainErr = gocbBucket.ExplainQuery(accessStatement, nil)
	assert.NoError(t, explainErr, "Error generating explain for access query")
	covered = isCovered(plan)
	planJSON, err = base.JSONMarshal(plan)
	assert.NoError(t, err)
	//assert.True(t, covered, "Access query isn't covered by index: %s", planJSON)

	// roleAccess
	roleAccessStatement := db.buildRoleAccessQuery("user1")
	plan, explainErr = gocbBucket.ExplainQuery(roleAccessStatement, nil)
	assert.NoError(t, explainErr, "Error generating explain for roleAccess query")
	covered = isCovered(plan)
	planJSON, err = base.JSONMarshal(plan)
	assert.NoError(t, err)
	//assert.True(t, !covered, "RoleAccess query isn't covered by index: %s", planJSON)

}

func TestAllDocsQuery(t *testing.T) {

	if base.UnitTestUrlIsWalrus() {
		t.Skip("This test is Couchbase Server only")
	}

	db, testBucket := setupTestDB(t)
	defer testBucket.Close()
	defer tearDownTestDB(t, db)

	// Add docs with channel assignment
	for i := 1; i <= 10; i++ {
		_, _, err := db.Put(fmt.Sprintf("allDocsTest%d", i), Body{"channels": []string{"ABC"}})
		assert.NoError(t, err, "Put allDocsTest doc")
	}

	// Standard query
	startKey := "a"
	endKey := ""
	results, queryErr := db.QueryAllDocs(startKey, endKey)
	assert.NoError(t, queryErr, "Query error")
	var row map[string]interface{}
	rowCount := 0
	for results.Next(&row) {
		rowCount++
	}
	assert.Equal(t, 10, rowCount)

	// Attempt to invalidate standard query
	startKey = "a' AND 1=0\x00"
	endKey = ""
	results, queryErr = db.QueryAllDocs(startKey, endKey)
	assert.NoError(t, queryErr, "Query error")
	rowCount = 0
	for results.Next(&row) {
		rowCount++
	}
	assert.Equal(t, 10, rowCount)

	// Attempt to invalidate statement to add row to resultset
	startKey = `a' UNION ALL SELECT TOSTRING(BASE64_DECODE("SW52YWxpZERhdGE=")) as id;` + "\x00"
	endKey = ""
	results, queryErr = db.QueryAllDocs(startKey, endKey)
	assert.NoError(t, queryErr, "Query error")
	rowCount = 0
	for results.Next(&row) {
		assert.NotEqual(t, row["id"], "InvalidData")
		rowCount++
	}
	assert.Equal(t, 10, rowCount)

	// Attempt to create syntax error
	startKey = `a'1`
	endKey = ""
	results, queryErr = db.QueryAllDocs(startKey, endKey)
	assert.NoError(t, queryErr, "Query error")
	rowCount = 0
	for results.Next(&row) {
		rowCount++
	}
	assert.Equal(t, 10, rowCount)

}

func TestAccessQuery(t *testing.T) {
	if base.UnitTestUrlIsWalrus() {
		t.Skip("This test is Couchbase Server only")
	}

	db, testBucket := setupTestDB(t)
	defer testBucket.Close()
	defer tearDownTestDB(t, db)

	db.ChannelMapper = channels.NewChannelMapper(`function(doc, oldDoc) {
	access(doc.accessUser, doc.accessChannel)
}`)
	// Add docs with access grants assignment
	for i := 1; i <= 5; i++ {
		_, _, err := db.Put(fmt.Sprintf("accessTest%d", i), Body{"accessUser": "user1", "accessChannel": fmt.Sprintf("channel%d", i)})
		assert.NoError(t, err, "Put accessTest doc")
	}

	// Standard query
	username := "user1"
	results, queryErr := db.QueryAccess(username)
	assert.NoError(t, queryErr, "Query error")
	var row map[string]interface{}
	rowCount := 0
	for results.Next(&row) {
		rowCount++
	}
	assert.Equal(t, 5, rowCount)

	// Attempt to introduce syntax error.  Should return zero rows for user `user1'`, and not return error
	username = "user1'"
	results, queryErr = db.QueryAccess(username)
	assert.NoError(t, queryErr, "Query error")
	rowCount = 0
	for results.Next(&row) {
		rowCount++
	}
	assert.Equal(t, 0, rowCount)

	// Attempt to introduce syntax error.  Should return zero rows for user `user1`AND`, and not return error.
	// Validates select clause protection
	username = "user1`AND"
	results, queryErr = db.QueryAccess(username)
	assert.NoError(t, queryErr, "Query error")
	rowCount = 0
	for results.Next(&row) {
		rowCount++
	}
	assert.Equal(t, 0, rowCount)
}

func TestRoleAccessQuery(t *testing.T) {
	if base.UnitTestUrlIsWalrus() {
		t.Skip("This test is Couchbase Server only")
	}

	db, testBucket := setupTestDB(t)
	defer testBucket.Close()
	defer tearDownTestDB(t, db)

	db.ChannelMapper = channels.NewChannelMapper(`function(doc, oldDoc) {
	role(doc.accessUser, "role:" + doc.accessChannel)
}`)
	// Add docs with access grants assignment
	for i := 1; i <= 5; i++ {
		_, _, err := db.Put(fmt.Sprintf("accessTest%d", i), Body{"accessUser": "user1", "accessChannel": fmt.Sprintf("channel%d", i)})
		assert.NoError(t, err, "Put accessTest doc")
	}

	// Standard query
	username := "user1"
	results, queryErr := db.QueryRoleAccess(username)
	assert.NoError(t, queryErr, "Query error")
	var row map[string]interface{}
	rowCount := 0
	for results.Next(&row) {
		rowCount++
	}
	assert.Equal(t, 5, rowCount)

	// Attempt to introduce syntax error.  Should return zero rows for user `user1'`, and not return error
	username = "user1'"
	results, queryErr = db.QueryRoleAccess(username)
	assert.NoError(t, queryErr, "Query error")
	rowCount = 0
	for results.Next(&row) {
		rowCount++
	}
	assert.Equal(t, 0, rowCount)

	// Attempt to introduce syntax error.  Should return zero rows for user `user1`AND`, and not return error
	// Validates select clause protection
	username = "user1`AND"
	results, queryErr = db.QueryRoleAccess(username)
	assert.NoError(t, queryErr, "Query error")
	rowCount = 0
	for results.Next(&row) {
		rowCount++
	}
	assert.Equal(t, 0, rowCount)
}

// Parse the plan looking for use of the fetch operation (appears as the key/value pair "#operator":"Fetch")
// If there's no fetch operator in the plan, we can assume the query is covered by the index.
// The plan returned by an EXPLAIN is a nested hierarchy with operators potentially appearing at different
// depths, so need to traverse the JSON object.
// https://docs.couchbase.com/server/6.0/n1ql/n1ql-language-reference/explain.html
func isCovered(plan map[string]interface{}) bool {
	for key, value := range plan {
		switch value := value.(type) {
		case string:
			if key == "#operator" && value == "Fetch" {
				return false
			}
		case map[string]interface{}:
			if !isCovered(value) {
				return false
			}
		case []interface{}:
			for _, arrayValue := range value {
				jsonArrayValue, ok := arrayValue.(map[string]interface{})
				if ok {
					if !isCovered(jsonArrayValue) {
						return false
					}
				}
			}
		default:
		}
	}
	return true
}

func countQueryResults(results sgbucket.QueryResultIterator) int {

	count := 0
	var row map[string]interface{}
	for results.Next(&row) {
		count++
	}
	return count
}

func TestQueryChannelsActiveOnlyWithLimit(t *testing.T) {

	if base.UnitTestUrlIsWalrus() {
		t.Skip("This test require Couchbase Server")
	}

	db, testBucket := setupTestDB(t)
	defer testBucket.Close()
	defer tearDownTestDB(t, db)

	// Create doc1, revision 1-a
	body := Body{"channels": []string{"ABC"}, "name": "Alice"}
	doc, _, err := db.PutExistingRevWithBody("doc1", body, []string{"1-a"}, false)
	require.NoError(t, err, "Couldn't create document")
	startSeq := doc.Sequence

	// Tombstone doc1, revision 2-a
	body[BodyDeleted] = true
	doc, _, err = db.PutExistingRevWithBody("doc1", body, []string{"2-a", "1-a"}, false)
	require.NoError(t, err, "Couldn't create document")

	// Create doc2 and two conflicting changes:
	body["name"] = "Bob"
	_, _, err = db.PutExistingRevWithBody("doc2", body, []string{"1-a"}, false)
	require.NoError(t, err, "Couldn't create document")

	body["name"] = "Owen"
	_, _, err = db.PutExistingRevWithBody("doc2", body, []string{"2-b", "1-a"}, false)
	require.NoError(t, err, "Couldn't create revision 2-b of doc2")

	body["name"] = "Chloe"
	_, _, err = db.PutExistingRevWithBody("doc2", body, []string{"2-a", "1-a"}, false)
	require.NoError(t, err, "Couldn't create revision 2-a of doc2")

	// Tombstone the non-winning branch of a conflicted document
	body[BodyDeleted] = true
	_, _, err = db.PutExistingRevWithBody("doc2", body, []string{"3-a", "2-a"}, false)
	assert.NoError(t, err, "Add 3-a (tombstone)")

	// Create doc3
	body["name"] = "Emily"
	doc, _, err = db.PutExistingRevWithBody("doc3", body, []string{"1-a"}, false)
	require.NoError(t, err, "Couldn't create revision 2-a of doc3")
	endSeq := doc.Sequence

	// Issue channel query with limit and activeOnly true
	entries, err := db.getChangesInChannelFromQuery("ABC", startSeq, endSeq, 2, true)
	assert.Len(t, entries, 1)

	// Issue channel query with no limit and activeOnly true
	entries, err = db.getChangesInChannelFromQuery("ABC", startSeq, endSeq, 0, true)
	assert.Len(t, entries, 3)

	// Issue channel query with limit and activeOnly false
	entries, err = db.getChangesInChannelFromQuery("ABC", startSeq, endSeq, 3, false)
	assert.Len(t, entries, 3)

	// Issue channel query with no limit and activeOnly false
	entries, err = db.getChangesInChannelFromQuery("ABC", startSeq, endSeq, 0, false)
	assert.Len(t, entries, 3)
}
