all:
	go build

query:
	go run bradgoodman.com/goprint  query http://10.0.0.202 MakeItLabel.ps

print:
	go run bradgoodman.com/goprint  print http://10.0.0.202 MakeItLabel.ps FirstLine SecondLine
