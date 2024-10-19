# Go Web Server

Generate Data Access Layer
```
sqlc generate
```

Generate Migrations
```bash
# Migration Up
 goose postgres "postgres://username:@localhost:5432/chirpy" up

 # Migration Down
 goose postgres "postgres://username:@localhost:5432/chirpy" down
```