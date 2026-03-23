for i in {1..5}
do
  go run cmd/cache/main.go $((8080 + i)) &
done