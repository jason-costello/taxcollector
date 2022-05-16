
create sequence "improvementDetail_id_seq"
    as integer;

alter sequence "improvementDetail_id_seq" owner to jc;

create sequence properties_id_seq
    as integer;

alter sequence properties_id_seq owner to jc;

create sequence rollvalues_id_seq
    as integer;

alter sequence rollvalues_id_seq owner to jc;

create sequence "main_roll_values_id_seq"
    as integer;

alter sequence "main_roll_values_id_seq" owner to jc;

create table jurisdictions
(
    id               serial
        constraint jurisdictions_pk
            primary key,
    entity           varchar(255),
    description      text,
    tax_rate        integer,
    appraised_value integer,
    taxable_value   integer,
    estimated_tax   integer,
    property_id     integer
);

alter table jurisdictions
    owner to jc;

create table land
(
    id            serial
        constraint land_pk
            primary key,
    number        integer,
    land_type          varchar(255),
    description   text,
    acres         double precision,
    square_feet  double precision,
    eff_front    double precision,
    eff_depth    double precision,
    market_value integer,
    property_id  integer
);

alter table land
    owner to jc;

create table properties
(
    id                    integer                not null
        constraint properties_pk
            primary key,
    owner_id             integer,
    owner_name           varchar(255),
    owner_mailing_address varchar(255),
    zoning                varchar(255),
    neighborhood_cd      varchar(255),
    neighborhood          varchar(500),
    address               varchar(500),
    legal_description    varchar(500),
    geographic_id        varchar(255),
    exemptions            varchar(255),
    ownership_percentage varchar(255),
    mapsco_map_id         varchar(255),
    longitude             numeric,
    latitude              numeric,
    address_number       varchar(255) default 0 not null,
    address_line_two      varchar(255),
    city                  varchar(255),
    street                varchar(255),
    county                varchar(255),
    state                 varchar(2)
);

alter table properties
    owner to jc;

alter sequence properties_id_seq owned by properties.id;

create table proxies
(
    ip       text not null
        constraint proxies_pk
            primary key,
    lastused text,
    uses     integer,
    is_bad  integer
);

alter table proxies
    owner to jc;

create table roll_values
(
    id             integer default nextval('"main_rollValues_id_seq"'::regclass) not null
        constraint main_rollvalues_pk
            primary key,
    year           integer,
    improvements   integer,
    land_market   integer,
    ag_valuation  integer,
    appraised      integer,
    homestead_cap integer,
    assessed       integer,
    property_id   integer
);

alter table roll_values
    owner to jc;

alter sequence "main_rollValues_id_seq" owned by roll_values.id;

create table improvement_detail
(
    id             integer default nextval('"improvementDetail_id_seq"'::regclass) not null
        constraint improvementdetail_pk
            primary key,
    improvement_id integer,
    improvement_type           varchar(255),
    description    text,
    class          varchar(255),
    exterior_wall varchar(255),
    year_built    integer,
    square_feet   integer
);

alter table improvement_detail
    owner to jc;

alter sequence "improvementDetail_id_seq" owned by improvement_detail.id;

create index improvementdetail_yearbuilt_index
    on improvement_detail ("yearBuilt");

create table improvements
(
    id           serial
        constraint improvements_pk
            primary key,
    name         text,
    description  text,
    state_code  varchar(255),
    living_area numeric,
    value        integer,
    property_id integer
);

alter table improvements
    owner to jc;

