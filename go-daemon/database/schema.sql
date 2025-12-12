CREATE TABLE IF NOT EXISTS containers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  container_id TEXT,
  pid INTEGER,
  image TEXT,
  cpu_pct REAL,
  mem_pct REAL,
  ts INTEGER
);

CREATE TABLE IF NOT EXISTS deletions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  container_id TEXT,
  reason TEXT,
  ts INTEGER
);

CREATE TABLE IF NOT EXISTS sys_metrics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  mem_total_kb INTEGER,
  mem_free_kb INTEGER,
  mem_used_kb INTEGER,
  ts INTEGER
);

CREATE TABLE IF NOT EXISTS process_count (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  total INTEGER,
  ts INTEGER
);
