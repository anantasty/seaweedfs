package weed_server

import (
	"context"
	"net/http"
	"time"

	"github.com/joeslay/seaweedfs/weed/operation"
	"github.com/joeslay/seaweedfs/weed/pb/volume_server_pb"
	"github.com/joeslay/seaweedfs/weed/storage/needle"
)

func (vs *VolumeServer) BatchDelete(ctx context.Context, req *volume_server_pb.BatchDeleteRequest) (*volume_server_pb.BatchDeleteResponse, error) {

	resp := &volume_server_pb.BatchDeleteResponse{}

	now := uint64(time.Now().Unix())

	for _, fid := range req.FileIds {
		vid, id_cookie, err := operation.ParseFileId(fid)
		if err != nil {
			resp.Results = append(resp.Results, &volume_server_pb.DeleteResult{
				FileId: fid,
				Status: http.StatusBadRequest,
				Error:  err.Error()})
			continue
		}

		n := new(needle.Needle)
		volumeId, _ := needle.NewVolumeId(vid)
		n.ParsePath(id_cookie)

		cookie := n.Cookie
		if _, err := vs.store.ReadVolumeNeedle(volumeId, n); err != nil {
			resp.Results = append(resp.Results, &volume_server_pb.DeleteResult{
				FileId: fid,
				Status: http.StatusNotFound,
				Error:  err.Error(),
			})
			continue
		}

		if n.IsChunkedManifest() {
			resp.Results = append(resp.Results, &volume_server_pb.DeleteResult{
				FileId: fid,
				Status: http.StatusNotAcceptable,
				Error:  "ChunkManifest: not allowed in batch delete mode.",
			})
			continue
		}

		if n.Cookie != cookie {
			resp.Results = append(resp.Results, &volume_server_pb.DeleteResult{
				FileId: fid,
				Status: http.StatusBadRequest,
				Error:  "File Random Cookie does not match.",
			})
			break
		}
		n.LastModified = now
		if size, err := vs.store.DeleteVolumeNeedle(volumeId, n); err != nil {
			resp.Results = append(resp.Results, &volume_server_pb.DeleteResult{
				FileId: fid,
				Status: http.StatusInternalServerError,
				Error:  err.Error()},
			)
		} else {
			resp.Results = append(resp.Results, &volume_server_pb.DeleteResult{
				FileId: fid,
				Status: http.StatusAccepted,
				Size:   size},
			)
		}
	}

	return resp, nil

}
