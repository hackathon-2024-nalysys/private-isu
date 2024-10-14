package main

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/catatsuy/private-isu/webapp/golang/types"
)

// returns hit users and missed ids
func getUsers(ids []int) (map[int]types.User, []int, error) {
	// convert to string array
	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = "user:" + strconv.Itoa(id)
	}

	hitsRaw, err := memcacheClient.GetMulti(strIDs)
	if err != nil {
		return nil, nil, err
	}

	hits := make(map[int]types.User, len(hitsRaw))

	// decode using gob
	var buffer bytes.Buffer
	for _, hitRaw := range hitsRaw {
		dec := gob.NewDecoder(&buffer)
		buffer.Write(hitRaw.Value)
		var hit types.User
		if err := dec.Decode(&hit); err != nil {
			return nil, nil, err
		}
		hits[hit.ID] = hit
	}

	missed := make([]int, 0, len(ids)-len(hits))
	for _, id := range ids {
		_, found := hitsRaw["user:"+strconv.Itoa(id)]
		if !found {
			missed = append(missed, id)
		}
	}

	return hits, missed, nil
}

// returns hit comments and missed ids
func getCommentsForPosts(postIDs []int) (map[int][]types.Comment, []int, error) {
	// convert to string array
	strIDs := make([]string, len(postIDs))
	for i, id := range postIDs {
		strIDs[i] = "post:" + strconv.Itoa(id)
	}

	hitsRaw, err := memcacheClient.GetMulti(strIDs)
	if err != nil {
		return nil, nil, err
	}

	hits := make(map[int][]types.Comment, len(hitsRaw))

	// decode using gob
	var buffer bytes.Buffer
	for _, hitRaw := range hitsRaw {
		dec := gob.NewDecoder(&buffer)
		buffer.Write(hitRaw.Value)
		var hit []types.Comment
		if err := dec.Decode(&hit); err != nil {
			return nil, nil, err
		}
		id, _ := strconv.Atoi(strings.Split(hitRaw.Key, ":")[1])
		hits[id] = hit
	}

	missed := make([]int, 0, len(postIDs)-len(hits))
	for _, id := range postIDs {
		_, found := hitsRaw["post:"+strconv.Itoa(id)]
		if !found {
			missed = append(missed, id)
		}
	}

	return hits, missed, nil
}

func cacheUsers(users []types.User) error {
	for _, user := range users {
		var buffer bytes.Buffer
		enc := gob.NewEncoder(&buffer)
		if err := enc.Encode(user); err != nil {
			return err
		}
		if err := memcacheClient.Set(&memcache.Item{
			Key:   "user:" + strconv.Itoa(user.ID),
			Value: buffer.Bytes(),
		}); err != nil {
			return err
		}
	}
	return nil
}

func cacheCommentsForPosts(postID int, comments []types.Comment) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(comments); err != nil {
		return err
	}
	if err := memcacheClient.Set(&memcache.Item{
		Key:   "post:" + strconv.Itoa(postID),
		Value: buffer.Bytes(),
	}); err != nil {
		return err
	}
	return nil
}

func invalidateUser(userID int) error {
	return memcacheClient.Delete("user:" + strconv.Itoa(userID))
}

func invalidateCommentsForPost(postID int) error {
	return memcacheClient.Delete("post:" + strconv.Itoa(postID))
}
