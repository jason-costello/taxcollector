-- query.sql

-- name: GetImprovementDetails :many
SELECT * FROM improvement_detail
WHERE improvement_id = $1;

-- name: GetRollValuesByPropertyID :many
Select * from roll_values
where property_id = $1;

-- name: IsExistingProperty :one
select exists(select 1 from properties where id = $1);

-- name: GetRandomURLs :many
SELECT url  FROM pending_urls
ORDER BY RANDOM()
LIMIT $1;

-- name: RemovePendingURL :exec
Delete from pending_urls where url = $1;

-- name: GetRemainingURLCount :one
Select count(url) from pending_urls;
-- name: GetImprovementDetail :one
SELECT * FROM improvement_detail
WHERE id = $1 LIMIT 1;

-- name: GetImprovementByID :one
SELECT * FROM improvements
WHERE id = $1 limit 1;

-- name: GetImprovementsByPropertyID :many
SELECT * FROM improvements
WHERE property_id = $1;

-- name: GetJurisdictionsByPropertyID :many
SELECT * FROM jurisdictions
WHERE property_id = $1;

-- name: GetValidProxy :one
select ip, lastused, uses
from proxies
where is_bad = 0
order by lastused asc, uses asc
limit 1;

-- name: UpdateProxyLastUsedTime :exec
update proxies set lastused = $1, uses = $2 where ip = $3;

-- name: InsertLand :exec
insert into land(number, land_type, description, acres, square_feet, eff_front, eff_depth, market_value, property_id) values($1,$2,$3,$4,$5,$6,$7,$8,$9);

-- name: InsertPropertyRecord :exec
insert into properties(id,owner_id,owner_name,owner_mailing_address,
                       zoning,neighborhood_cd,neighborhood,
                       address, legal_description, geographic_id, exemptions,
                       ownership_percentage, mapsco_map_id)
values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13);

-- name: InsertRollValue :exec
insert into roll_values( year, improvements, land_market, ag_valuation, appraised, homestead_cap, assessed, property_id) values($1,$2,$3,$4,$5,$6,$7,$8);

-- name: InsertJurisdiction :exec
insert into jurisdictions( entity, description, tax_rate, appraised_value, taxable_value, estimated_tax, property_id) values($1,$2,$3,$4,$5,$6,$7);

-- name: InsertImprovement :one
insert into improvements (name, description, state_code, living_area, value, property_id) values($1,$2,$3,$4,$5,$6) RETURNING id;;

-- name: InsertImprovementDetail :exec
insert into improvement_detail(improvement_id, improvement_type, description, class, exterior_wall, year_built, square_feet) values ($1,$2,$3,$4,$5,$6,$7) ;


-- name: GetLandByPropertyID :many
SELECT * FROM land
WHERE property_id = $1;

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

-- name: GetPropertyByStreet :many
Select * from properties where UPPER(street) = UPPER($1) order by address_number,street,city asc;


-- name: ListProperties :many
Select * from properties limit $1 offset $2;

-- name: UpdatePropertySetAddressParts :exec
Update properties set address_number = $1, address_line_two = $2, street = $3, city = $4, county = $5, state = $6
where id = $7;

-- name: GetStreetsLike :many
Select  distinct street from properties where street like concat($1::text,'%') order by street asc;

-- name: GetNeighborhoodsLike :many
Select  distinct neighborhood from properties where Upper(neighborhood) like concat(Upper($1)::text,'%') order by neighborhood asc;


-- name: GetDistinctStreets :many
Select Distinct street from properties order by street asc;



-- name: GetDistinctNeighborhoods :many
Select Distinct neighborhood from properties order by neighborhood asc;
