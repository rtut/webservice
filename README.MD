## Installation and startup instructions for go wev service
1. Install PostgreSQL 9.6.10
2. Install golang 1.10
3. Application of migrations to the database
   * user@user$ sudo psql -U postgres  (connect to postgres)
   * postgres=# \i path_to_/migration_1.sql    (migration_1.sql locate in project in folder sql)
   * test_app=# \q
4. Start application  
   * user@user$ cd path/to/project/folder/webservice  
   * user@user$ ./webservice
   
## API info

* ### Add group  

    Method: PUT   
    http://0.0.0.0:5555/group/add  
    Json: {  
        "name": "new_name",
        "parent_group_id": number parent  
        }
    
    If you add group to root, you must send 0 in parent_group_id in json

* ### Get group 
        
    Method: GET                                             
    http://0.0.0.0:5555/group/get                                          
    Json: {                                                             
        "id": number group                                              
        }                                    

* ### Get tree group
     
    Method: GET                                                         
    http://0.0.0.0:5555/group/get/tree                                          
    Json: {                                                             
        "id": number group,                                             
        "depth": number  
        }                                    
    
    If the depth exceeds the actual, you will be given all the 
    groups with all the nested subgroups                                                               
    
* ### Move group

    Method: PATCH                                                         
    http://0.0.0.0:5555/group/move                                         
    Json: {                                                             
        "name": "name group",                                             
        "parent_group_d": number,
        "new_parent_group_id": number  
        }                                    
    
    Moves the group to the specified subgroup, if you specify 0 in 
    new_parent_group_id, the group with all its nesting will move to the root                                                              
    
* ### Rename group
    
    Method: PATCH                                                         
    http://0.0.0.0:5555/group/rename                                         
    Json: {                                                             
        "name": "old_name_group",                                             
        "parent_group_d": number,
        "new_name": "new_name_group"  
        }                                    
    
* ### Delete group
     
    Method: DELETE                                                         
    http://0.0.0.0:5555/group/delete   
    Json: {                                                             
        "id": number group,                                             
        }                                    
    
    Delete group with all subgroups 

