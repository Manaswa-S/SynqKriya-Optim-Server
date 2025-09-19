

-- name: GetCameraInfo :one
SELECT
    cameras.camera_id,
    cameras.junction_id,
    cameras.rtsp_url,
    cameras.angle,
    cameras.resolution,
    cameras.status::TEXT AS status,
    cameras.created_at,
    junctions.name AS juncName,
    junctions.latitude,
    junctions.longitude
FROM cameras
JOIN junctions ON junctions.junction_id = cameras.junction_id
WHERE cameras.camera_id = $1;


-- name: GetAllCameras :many
SELECT
    cameras.camera_id,
    cameras.junction_id,
    cameras.rtsp_url,
    cameras.angle,
    cameras.resolution,
    cameras.status::TEXT AS status,
    cameras.created_at,
    junctions.name AS juncName,
    junctions.latitude,
    junctions.longitude
FROM cameras
JOIN junctions ON junctions.junction_id = cameras.junction_id;    


-- name: GetPoliciesForCameraID :one
WITH temp AS (
    SELECT
        cameras.policy_id
    FROM cameras
    WHERE cameras.camera_id = $1
)

SELECT
    policies.policy_id,
    policies.green_min,
    policies.green_max,
    policies.red_min,
    policies.red_max,
    policies.cycle_max
FROM policies
WHERE policies.policy_id = temp.policy_id;