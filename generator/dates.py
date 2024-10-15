import datetime
import random

def generate_commit_dates(count=180, start_date=datetime.date(2025, 10, 6), end_date=datetime.date(2026, 5, 20)):
    # Seed for deterministic timestamps
    rng = random.Random(42)
    
    current_date = start_date
    delta_days = (end_date - start_date).days
    
    dates = []
    # Generate weighted days
    day_pool = []
    d = start_date
    while d <= end_date:
        # Weekdays get weight 5, weekends get weight 1
        if d.weekday() < 5:
            day_pool.extend([d] * 5)
        else:
            day_pool.extend([d] * 1)
        d += datetime.timedelta(days=1)
        
    selected_days = sorted(rng.sample(day_pool, count))
    
    last_dt = None
    for day in selected_days:
        hour = rng.choice([9, 10, 11, 13, 14, 15, 16, 17, 18, 19, 20])
        minute = rng.randint(0, 59)
        second = rng.randint(0, 59)
        dt = datetime.datetime(day.year, day.month, day.day, hour, minute, second, tzinfo=datetime.timezone.utc)
        if last_dt and dt <= last_dt:
            dt = last_dt + datetime.timedelta(minutes=rng.randint(12, 45), seconds=rng.randint(5, 55))
        last_dt = dt
        dates.append(dt.strftime("%Y-%m-%dT%H:%M:%SZ"))
        
    return dates

if __name__ == '__main__':
    d = generate_commit_dates(10)
    print("Sample dates:", d)
