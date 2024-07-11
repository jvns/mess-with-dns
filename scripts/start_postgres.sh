# forwards the port 5432 to 5432
export DOCKER_HOST=unix:///var/run/docker.sock
docker run -p 5432:5432 --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -d postgres
