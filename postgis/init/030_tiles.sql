-- Vector tile function served by Martin as the "activities" source.
--
-- SECURITY: tiles only ever contain the activities of the user given in
-- query_params->>'user_id'. Martin is not reachable from outside the
-- compose network; the backend authenticates the request and sets user_id
-- itself, ignoring whatever the browser sent. No user_id means no tile.

CREATE FUNCTION activities_tiles(z integer, x integer, y integer, query_params json)
RETURNS bytea
LANGUAGE plpgsql
STABLE
PARALLEL SAFE
SECURITY DEFINER
SET search_path = public, pg_temp
AS $$
DECLARE
    uid  uuid;
    bbox geometry := ST_TileEnvelope(z, x, y);
    mvt  bytea;
BEGIN
    BEGIN
        uid := (query_params ->> 'user_id')::uuid;
    EXCEPTION WHEN invalid_text_representation THEN
        RETURN NULL;
    END;
    IF uid IS NULL THEN
        RETURN NULL;
    END IF;

    SELECT ST_AsMVT(t, 'activities', 4096, 'geom')
      INTO mvt
      FROM (
        SELECT a.id::text                              AS id,
               a.name,
               a.sport,
               extract(epoch FROM a.started_at)::bigint AS started_at,
               round(a.distance_m)::integer            AS distance_m,
               ST_AsMVTGeom(
                   CASE
                       WHEN z < 10 THEN a.geom_lo
                       WHEN z < 13 THEN a.geom_mid
                       ELSE a.geom_3857
                   END,
                   bbox, 4096, 64, true)               AS geom
          FROM activities a
         WHERE a.user_id = uid
           AND a.geom_3857 && ST_Expand(bbox, (ST_XMax(bbox) - ST_XMin(bbox)) * 64 / 4096)
      ) t
     WHERE t.geom IS NOT NULL;

    RETURN mvt;
END;
$$;

REVOKE ALL ON FUNCTION activities_tiles(integer, integer, integer, json) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION activities_tiles(integer, integer, integer, json) TO mytracks_martin;
