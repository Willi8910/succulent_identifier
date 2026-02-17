# Database Setup Guide

## PostgreSQL Database for Succulent Identifier

### Quick Start with Docker Compose

The easiest way to run PostgreSQL is using Docker Compose:

```bash
# From project root
docker-compose up -d postgres

# Check if PostgreSQL is running
docker-compose ps

# View logs
docker-compose logs postgres
```

This will start PostgreSQL on `localhost:5432` with:
- Database: `succulent_identifier`
- User: `postgres`
- Password: `postgres`

### Manual PostgreSQL Installation

If you prefer to install PostgreSQL manually:

**macOS (Homebrew)**:
```bash
brew install postgresql@16
brew services start postgresql@16
createdb succulent_identifier
```

**Ubuntu/Debian**:
```bash
sudo apt-get install postgresql-16
sudo systemctl start postgresql
sudo -u postgres createdb succulent_identifier
```

**Windows**:
Download from https://www.postgresql.org/download/windows/

### Database Schema

The application automatically creates the following tables on startup:

#### **identifications** table
Stores plant identification records:
```sql
id              UUID PRIMARY KEY
genus           VARCHAR(255)
species         VARCHAR(255)
confidence      DECIMAL(5,4)
image_path      TEXT
created_at      TIMESTAMP
```

#### **chat_messages** table
Stores chat conversation history:
```sql
id                  UUID PRIMARY KEY
identification_id   UUID (FK → identifications.id)
message             TEXT
sender              VARCHAR(10) ('user' or 'llm')
created_at          TIMESTAMP
```

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cd backend
cp .env.example .env
```

Edit `.env`:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=succulent_identifier
DB_SSLMODE=disable
```

### Running Migrations

Migrations run automatically when the backend starts. To run manually:

```bash
cd backend
go run main.go
```

### Verify Database Setup

Connect to PostgreSQL:
```bash
# Using psql
psql -h localhost -U postgres -d succulent_identifier

# List tables
\dt

# Check identifications table
SELECT * FROM identifications;

# Check chat_messages table
SELECT * FROM chat_messages;
```

### Stopping the Database

```bash
# Stop PostgreSQL container
docker-compose down

# Stop and remove data
docker-compose down -v
```

### Troubleshooting

**Port 5432 already in use**:
```bash
# Check what's using port 5432
lsof -i:5432

# Stop other PostgreSQL instances
brew services stop postgresql
# or
sudo systemctl stop postgresql
```

**Connection refused**:
- Ensure PostgreSQL is running: `docker-compose ps`
- Check logs: `docker-compose logs postgres`
- Verify port: `lsof -i:5432`

**Permission denied**:
- Check DB_USER and DB_PASSWORD in .env
- Ensure user has correct permissions

### Database Backup

**Backup**:
```bash
docker-compose exec postgres pg_dump -U postgres succulent_identifier > backup.sql
```

**Restore**:
```bash
docker-compose exec -T postgres psql -U postgres succulent_identifier < backup.sql
```

### Production Considerations

For production deployment:

1. **Use strong passwords**:
   ```env
   DB_PASSWORD=<strong-random-password>
   ```

2. **Enable SSL**:
   ```env
   DB_SSLMODE=require
   ```

3. **Use managed database** (AWS RDS, Google Cloud SQL, etc.)

4. **Set connection pooling** in `db/database.go`:
   ```go
   DB.SetMaxOpenConns(25)
   DB.SetMaxIdleConns(5)
   DB.SetConnMaxLifetime(5 * time.Minute)
   ```

5. **Enable backups** (automated daily backups)

6. **Monitor performance** (slow query logs, connection stats)

---

**Database setup complete!** ✅

The backend will automatically:
- Connect to PostgreSQL on startup
- Run migrations to create tables
- Be ready to save identification and chat history
