package main

import (
	"strconv"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/catatsuy/private-isu/webapp/golang/grpc"
	"github.com/catatsuy/private-isu/webapp/golang/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	// decode using grpc
	for _, hitRaw := range hitsRaw {
		var user grpc.User
		err := proto.Unmarshal(hitRaw.Value, &user)
		if err != nil {
			return nil, nil, err
		}
		hits[int(user.ID)] = types.User{
			ID:          int(user.ID),
			AccountName: user.AccountName,
			Authority:   int(user.Authority),
			DelFlg:      int(user.DelFlg),
			CreatedAt:   user.CreatedAt.AsTime(),
		}
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

	// decode using grpc
	for _, hitRaw := range hitsRaw {
		var comments grpc.Comments
		err := proto.Unmarshal(hitRaw.Value, &comments)
		if err != nil {
			return nil, nil, err
		}
		hit := make([]types.Comment, len(comments.Comments))
		for i, comment := range comments.Comments {
			hit[i] = types.Comment{
				ID:        int(comment.ID),
				PostID:    int(comment.PostID),
				UserID:    int(comment.UserID),
				Comment:   comment.Comment,
				CreatedAt: comment.CreatedAt.AsTime(),
				User: types.User{
					ID:          int(comment.User.ID),
					AccountName: comment.User.AccountName,
					Authority:   int(comment.User.Authority),
				},
			}
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
		pb, err := proto.Marshal(&grpc.User{
			ID:          int32(user.ID),
			AccountName: user.AccountName,
			Authority:   int32(user.Authority),
			DelFlg:      int32(user.DelFlg),
			CreatedAt:   timestamppb.New(user.CreatedAt),
		})
		if err != nil {
			return err
		}
		if err := memcacheClient.Set(&memcache.Item{
			Key:   "user:" + strconv.Itoa(user.ID),
			Value: pb,
		}); err != nil {
			return err
		}
	}
	return nil
}

func cacheCommentsForPosts(postID int, comments []types.Comment) error {
	gc := grpc.Comments{
		Comments: make([]*grpc.Comment, len(comments)),
	}

	for i, comment := range comments {
		gc.Comments[i] = &grpc.Comment{
			ID:        int32(comment.ID),
			PostID:    int32(comment.PostID),
			UserID:    int32(comment.UserID),
			Comment:   comment.Comment,
			CreatedAt: timestamppb.New(comment.CreatedAt),
			User: &grpc.User{
				ID:          int32(comment.User.ID),
				AccountName: comment.User.AccountName,
				Authority:   int32(comment.User.Authority),
			},
		}
	}

	pb, err := proto.Marshal(&gc)
	if err != nil {
		return err
	}
	if err := memcacheClient.Set(&memcache.Item{
		Key:   "post:" + strconv.Itoa(postID),
		Value: pb,
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
