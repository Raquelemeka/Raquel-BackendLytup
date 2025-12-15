My code works by


Differences from main code:

- no complex channel-based daemon code
- variable names
- the daemon struct (simplified mine)
- simpler shutdown logic

My Code:

- simple daemon struct
- simple goroutines (is this fine?)
- only SIGINT and SIGTERM 
- integrated HTTP with Daemon (it was more intuitive for me to do it this way)


To test:

1. Build the project
go build -o sr05-service


2. Run it
./sr05-service

3. In another terminal, test it
curl -X POST http://localhost:8080/service \
  -H "Content-Type: application/json" \
  -d '{"Srvc":"sr05"}'


The output should be 
{
    "ExctnOutcomeCode": 200,
    "ExctnOutcomeNote": "",
    "Yield": "Hello world"
}