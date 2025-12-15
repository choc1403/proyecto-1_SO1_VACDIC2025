SELECT mem_total_kb / 1024.0 / 1024.0 AS "TOTAL DE RAM"
FROM sys_metrics
ORDER BY ts DESC
LIMIT 1;

SELECT mem_free_kb / 1024.0 / 1024.0 AS "MEMORIA RAM LIBRE"
FROM sys_metrics
ORDER BY ts DESC
LIMIT 1;

SELECT COUNT(*) AS "TOTAL DE CONTENEDORES ELIMINADOS"
FROM deletions;

SELECT 
  ts * 1000 AS time,
  mem_used_kb / 1024.0 / 1024.0 AS "USO DE RAM EN EL TIEMPO"
FROM sys_metrics
ORDER BY ts ASC;

-- TOP 5 +RAM
SELECT
  container_id,
  image,
  pid,
  cpu_pct,
  mem_pct,
  datetime(ts, 'unixepoch') AS timestamp
FROM containers
WHERE container_id IS NOT NULL AND container_id != ''
ORDER BY cpu_pct DESC
LIMIT 5;


-- TOP 5 +CPU
SELECT
  container_id,
  image,
  pid,
  mem_pct,
  cpu_pct,
  datetime(ts, 'unixepoch') AS timestamp
FROM containers
WHERE container_id IS NOT NULL AND container_id != ''
ORDER BY mem_pct DESC
LIMIT 5;


SELECT mem_used_kb / 1024.0 / 1024.0 AS "MEMORIA RAM USADA"
FROM sys_metrics
ORDER BY ts DESC
LIMIT 1;