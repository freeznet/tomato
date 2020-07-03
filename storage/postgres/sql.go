package postgres

// Function to set a key on a nested JSON document
const jsonObjectSetKey = `CREATE OR REPLACE FUNCTION "json_object_set_key"(
  "json"          jsonb,
  "key_to_set"    TEXT,
  "value_to_set"  anyelement
)
  RETURNS jsonb 
  LANGUAGE sql 
  IMMUTABLE 
  STRICT 
AS $function$
SELECT concat('{', string_agg(to_json("key") || ':' || "value", ','), '}')::jsonb
  FROM (SELECT *
          FROM jsonb_each("json")
         WHERE "key" <> "key_to_set"
         UNION ALL
        SELECT "key_to_set", to_json("value_to_set")::jsonb) AS "fields"
$function$`

const arrayAdd = `CREATE OR REPLACE FUNCTION "array_add"(
  "array"   jsonb,
  "values"  jsonb
)
  RETURNS jsonb 
  LANGUAGE sql 
  IMMUTABLE 
  STRICT 
AS $function$ 
  SELECT array_to_json(ARRAY(SELECT unnest(ARRAY(SELECT DISTINCT jsonb_array_elements("array")) ||  ARRAY(SELECT jsonb_array_elements("values")))))::jsonb;
$function$`

const arrayAddUnique = `CREATE OR REPLACE FUNCTION "array_add_unique"(
  "array"   jsonb,
  "values"  jsonb
)
  RETURNS jsonb 
  LANGUAGE sql 
  IMMUTABLE 
  STRICT 
AS $function$ 
  SELECT array_to_json(ARRAY(SELECT DISTINCT unnest(ARRAY(SELECT DISTINCT jsonb_array_elements("array")) ||  ARRAY(SELECT DISTINCT jsonb_array_elements("values")))))::jsonb;
$function$`

const arrayRemove = `CREATE OR REPLACE FUNCTION "array_remove"(
  "array"   jsonb,
  "values"  jsonb
)
  RETURNS jsonb 
  LANGUAGE sql 
  IMMUTABLE 
  STRICT 
AS $function$ 
  SELECT array_to_json(ARRAY(SELECT * FROM jsonb_array_elements("array") as elt WHERE elt NOT IN (SELECT * FROM (SELECT jsonb_array_elements("values")) AS sub)))::jsonb;
$function$`

const arrayContainsAll = `CREATE OR REPLACE FUNCTION "array_contains_all"(
  "array"   jsonb,
  "values"  jsonb
)
  RETURNS boolean 
  LANGUAGE sql 
  IMMUTABLE 
  STRICT 
AS $function$ 
  SELECT RES.CNT = jsonb_array_length("values") FROM (SELECT COUNT(*) as CNT FROM jsonb_array_elements("array") as elt WHERE elt IN (SELECT jsonb_array_elements("values"))) as RES ;
$function$`

const arrayContains = `CREATE OR REPLACE FUNCTION "array_contains"(
  "array"   jsonb,
  "values"  jsonb
)
  RETURNS boolean 
  LANGUAGE sql 
  IMMUTABLE 
  STRICT 
AS $function$ 
  SELECT RES.CNT >= 1 FROM (SELECT COUNT(*) as CNT FROM jsonb_array_elements("array") as elt WHERE elt IN (SELECT jsonb_array_elements("values"))) as RES ;
$function$`

const timeBucketTimeZ = `CREATE OR REPLACE FUNCTION time_bucket_timez (
  bucket_width interval,
  ts timestamptz,
  "offset" interval = '00:00:00'::interval,
  origin timestamptz = '0001-01-01 00:00:00+00'::timestamptz
)
RETURNS TIMESTAMPTZ AS
$body$
/*
millenium = 1000 years
century = 100 years
decade = 10 years
1.5 years aka 1 year 6 months or 18 months
year aka 12 months
half year aka 6 months
quarter aka 3 months
months = months
*/
DECLARE
  months integer;
  bucket_months integer;
  bucket_month integer;
BEGIN
	IF EXTRACT(MONTH FROM bucket_width) >= 1 OR EXTRACT(YEAR FROM bucket_width) >= 1 THEN
    origin := origin + ((0-date_part('timezone_hour', now()))::text || ' hours')::interval;
      bucket_months :=
          (EXTRACT(MONTH FROM bucket_width) + -- months
          (EXTRACT(YEAR FROM bucket_width) * 12)); -- years
      months := (((EXTRACT(YEAR FROM ts)-EXTRACT(YEAR FROM origin))*12)+EXTRACT(MONTH FROM ts)-1);
      bucket_month := floor(months/bucket_months)*bucket_months;

      RETURN make_timestamptz(
          (EXTRACT(YEAR FROM origin)+floor(bucket_month/12))::integer, -- year
          (bucket_month%12)+1, -- month
          1, --day
          0, -- hour
          0, -- minute
          0, --second
          'Z');
	ELSE
      CASE
      WHEN "offset" > '0s'::interval THEN
	      RETURN public.time_bucket(bucket_width, ts-"offset") + "offset";
      WHEN origin <> '0001-01-01T00:00:00Z'::timestamptz THEN
	      RETURN public.time_bucket(bucket_width, ts, origin);
      ELSE
	      RETURN public.time_bucket(bucket_width, ts);
      END CASE;
    END IF;
END;
$body$
LANGUAGE 'plpgsql'
IMMUTABLE
RETURNS NULL ON NULL INPUT
SECURITY INVOKER;`