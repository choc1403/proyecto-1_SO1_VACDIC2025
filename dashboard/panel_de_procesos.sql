SELECT mem_total_kb / 1024.0 / 1024.0 AS "TOTAL DE RAM"
FROM sys_metrics
ORDER BY ts DESC
LIMIT 1;

SELECT mem_free_kb / 1024.0 / 1024.0 AS "MEMORIA RAM LIBRE"
FROM sys_metrics
ORDER BY ts DESC
LIMIT 1;

SELECT total AS "TOTAL DE PROCESOS CONTADOS"
FROM process_count
ORDER BY ts DESC
LIMIT 1;
SELECT 
  ts * 1000 AS time,
  mem_used_kb / 1024.0 / 1024.0 AS "USO DE RAM EN EL TIEMPO"
FROM sys_metrics
ORDER BY ts ASC;

-- TOP 5 +RAM
SELECT
  pid,
  image AS process_name,
  cpu_pct,
  mem_pct,
  datetime(ts, 'unixepoch') AS timestamp
FROM containers
WHERE container_id = '' OR container_id IS NULL
ORDER BY cpu_pct DESC
LIMIT 5;


-- TOP 5 +CPU
SELECT
  pid,
  image AS process_name,
  mem_pct,
  cpu_pct,
  datetime(ts, 'unixepoch') AS timestamp
FROM containers
WHERE container_id = '' OR container_id IS NULL
ORDER BY mem_pct DESC
LIMIT 5;


SELECT mem_used_kb / 1024.0 / 1024.0 AS "MEMORIA RAM USADA"
FROM sys_metrics
ORDER BY ts DESC
LIMIT 1;