-- query.sql


-- name: GetImprovementDetails :many
SELECT * FROM improvement_detail
WHERE improvement_id = $1;

-- name: GetRollValuesByPropertyID :many
Select * from roll_values
where property_id = $1;


-- name: GetImprovementDetail :one
SELECT * FROM improvement_detail
WHERE id = $1 LIMIT 1;

-- name: GetImprovementByID :one
SELECT * FROM improvements
WHERE id = $1 limit 1;

-- name: GetImprovementsByPropertyID :many
SELECT * FROM improvements
WHERE property_id = $1 limit 1;

-- name: GetJurisdictionsByPropertyID :many
SELECT * FROM jurisdictions
WHERE property_id = $1;

-- name: GetLandByPropertyID :many
SELECT * FROM land
WHERE property_id = $1 limit 1;

-- name: GetLandBySize :many
SELECT * FROM land
WHERE acres >= $1
 and acres <= $2;

-- name: GetLandByType :many
SELECT * FROM land
WHERE land_type = $1;

-- name: GetPropertyByID :one
SELECT * FROM properties
WHERE id = $1 limit 1;

-- name: GetPropertyByNeighborhood :many
SELECT * FROM properties
WHERE neighborhood = $1;

-- name: ListProperties :many
Select * from properties limit $1 offset $2;

-- name: UpdatePropertySetAddressParts :exec
Update properties set address_number = $1, address_line_two = $2, street = $3, city = $4, county = $5, state = $6
where id = $7;