- sigstp (ctrl+z) causing mongorestd to quit
- add sort, limit and skip query string parameters
- make the mode constants string (ro, rw, wo)
- add usage statement to mongorestd
	add it to readme as well
- add config file to mongorestd
- implement delete all
	think this needs to be added to rest.go
- makes sure the id in matches id out on put
- implement looking up by dates
- implement tests
	tests are out for now, don't know how to mock the database and web server

DONE:
- bug: objectIds being returned from post in wrong format
- add "event handlers"
- create resource struct
- add readonly mode
- added setting up resources with command line arguments
- added logging to a file with command line flag
- add logging 
- content negotiation for Index
- content negotiation for Find
- make test script
- add content-type handling for posts and puts
- apparently PUT should be an upsert not an insert (believe PUT should also fully replace the document, not just partially updated it -- verify)
- add basic type handling to querying
- add more complex methods
	* implemented find with simple querystring
- implement looking up by numbers
- figure out how to eliminate the struct required for decoding objects
- add content-type sending and accept receiving
- better error handling
- patch rest.go and send it in
- add all basic methods
- figure out how to make it generic
- turn from program to library

