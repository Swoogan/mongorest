#!/bin/sh
###########
### PUT ###
###########
#echo "Put"
#curl -i -H "Content-type: application/json" -X PUT -d '{"fname": "jim","lname": "bean"}' http://localhost:8080/employees/123
#
#echo "\nGet"
#curl -i -H "Accept: application/json" http://localhost:8080/employees/123
#
#echo "\nReplace"
#curl -i -H "Content-type: application/json" -X PUT -d '{"fname": "jimmy"}' http://localhost:8080/employees/123
#
#echo "\nGet"
#curl -i -H "Accept: application/json" http://localhost:8080/employees/123
#
#echo "\nDelete"
#curl -i -H "Accept: application/json" -X DELETE http://localhost:8080/employees/123

###################
### PUT WITH ID ### 
###################

#echo "Put with id"
#curl -i -H "Content-type: application/json" -X PUT -d '{"_id": 123, "fname": "jim","lname": "bean"}' http://localhost:8080/employees/123
#
#echo "\nGet"
#curl -i -H "Accept: application/json" http://localhost:8080/employees/123
#
#echo "\nDelete"
#curl -i -H "Accept: application/json" -X DELETE http://localhost:8080/employees/123

############
### POST ###
############
echo "Post with id"
curl -i -H "Content-type: application/json" -X POST -d '{"_id": 321, "fname": "johnny","lname": "test"}' http://localhost:8080/employees/ 

echo "\nGet"
curl -i -H "Accept: application/json" http://localhost:8080/employees/

echo "\nPartial Updated"
curl -i -H "Content-type: application/json" -X POST -d '{"_id": 321, "fname": "john"}' http://localhost:8080/employees/ 

echo "\nGet"
curl -i -H "Accept: application/json" http://localhost:8080/employees/
