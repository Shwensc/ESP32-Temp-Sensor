import sqlite3
from datetime import datetime, timedelta
from fastapi import FastAPI
from pydantic import BaseModel
from contextlib import asynccontextmanager

DB_PATH = "alerts.db"
ALERT_INTERVAL = 60

class TemperatureData(BaseModel):
    temperature: float

@asynccontextmanager
async def lifespan(_: FastAPI):
    # This runs at startup
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS alerts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            timestamp TEXT
        )
    """)
    conn.commit()
    conn.close()

    yield  # Run the app

    # This would run at shutdown (if needed)
    # (nothing to clean up here)

app = FastAPI(lifespan=lifespan)

@app.post("/alert")
async def check_temperature(data: TemperatureData):
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    cursor.execute("SELECT timestamp FROM alerts ORDER BY id DESC LIMIT 1")
    row = cursor.fetchone()

    now = datetime.now()

    if row:
        last_alert_time = datetime.fromisoformat(row[0])
        if now - last_alert_time < timedelta(seconds=ALERT_INTERVAL):
            print(f"[{now}] Alert suppressed. Last alert was at {last_alert_time}.")
            conn.close()
            return {"status": "Alert suppressed", "last_alert": last_alert_time.isoformat()}

    # Log new alert
    cursor.execute("INSERT INTO alerts (timestamp) VALUES (?)", (now.isoformat(),))
    conn.commit()
    conn.close()

    if data.temperature >= 30:
        print(f"[{now}] High Temperature:", data.temperature)
        return {"status": "High"}
    else:
        print(f"[{now}] Low Temperature:", data.temperature)
        return {"status": "Low"}

