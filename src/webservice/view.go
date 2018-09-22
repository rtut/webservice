package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

const (
	msgInternalError  = "Sorry, we have a problem on server, try again later."
	getId             = "SELECT * FROM groups WHERE id = $1;"
	addGroup          = "INSERT INTO groups (name, parent_group_id) VALUES ($1, $2) RETURNING id, name, parent_group_id;"
	loopGroup         = "SELECT id FROM groups WHERE name = $1 AND parent_group_id = $2;"
	duplicateGroup    = "SELECT EXISTS(SELECT 1 FROM groups where name = $1 and parent_group_id = $2)"
	existGroup        = "SELECT EXISTS(SELECT 1 from groups where id = $1);"
	updateNameGroup   = "UPDATE groups SET name = $1 where name = $2 and parent_group_id = $3 RETURNING id, name, parent_group_id;"
	updateParentGroup = "UPDATE groups SET parent_group_id = $1 where name = $2 and parent_group_id = $3 RETURNING id, name, parent_group_id;"
	// use recursive search, but could use nested set or may be through closure with triggers in db:D,
	// but this is a completely different story:wq
	selectTreeGroup = `with recursive all_info as (
							select id, name, parent_group_id, 1 as level from groups where id = $1

							union all

							select groups.id, groups.name, groups.parent_group_id, all_info.level + 1 as level
							from groups
								join all_info
									on groups.parent_group_id = all_info.id
						)

						select id, name, parent_group_id from all_info where level <= $2;`
	deleteTreeGroup = `with recursive all_info as (
							select id from groups where id = $1

							union all

							select groups.id from groups
								join all_info
									on groups.parent_group_id = all_info.id
						)

						delete from groups where id in (select id from all_info);`
)

type Groups struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	ParentGroupId int    `json:"parent_group_id"`
}

type NewNameGroup struct {
	Groups
	NewName string `json:"new_name"`
}

type NewParentGroup struct {
	Groups
	NewParentGroupId int `json:"new_parent_group_id"`
}

type RequestId struct {
	Id int `json:"id"`
}

type GetTree struct {
	RequestId
	Depth int `json:"depth"`
}

type Msg struct {
	Msg string `json:"msg"`
}

var msg Msg

func SetNameGroup(e echo.Context) error {
	var isExist bool
	renameGroup := NewNameGroup{}
	group := Groups{}
	err := e.Bind(&renameGroup)
	if err != nil {
		LocalLog.Println(err)
	}
	err = DB.QueryRow(duplicateGroup, renameGroup.NewName, renameGroup.ParentGroupId).Scan(&isExist)
	if err != nil {
		LocalLog.Println(err)
	}
	if isExist {
		msg := Msg{}
		msg.Msg = fmt.Sprintf("Group with this name %s already created in the database", renameGroup.NewName)
		return e.JSON(http.StatusBadRequest, msg)
	}

	// add info about new group in db
	err = DB.QueryRow(
		updateNameGroup, renameGroup.NewName, renameGroup.Name, renameGroup.ParentGroupId).Scan(&group.Id, &group.Name, &group.ParentGroupId)
	switch {
	case err == sql.ErrNoRows:
		return e.JSON(http.StatusNotFound, nil)
	case err != nil:
		LocalLog.Println(err)
		msg.Msg = msgInternalError
		return e.JSON(http.StatusInternalServerError, msg)
	}
	return e.JSON(http.StatusOK, group)
}

func GetGroup(e echo.Context) error {
	selectId := RequestId{}
	group := Groups{}
	err := e.Bind(&selectId)
	if err != nil {
		LocalLog.Println(err)
	}
	if (RequestId{}) == selectId {
		msg.Msg = "Check the data in your request."
		return e.JSON(http.StatusBadRequest, msg)
	}
	err = DB.QueryRow(getId, selectId.Id).Scan(
		&group.Id, &group.Name, &group.ParentGroupId)
	switch {
	case err == sql.ErrNoRows:
		return e.JSON(http.StatusNotFound, nil)
	case err != nil:
		LocalLog.Println(err)
		msg.Msg = msgInternalError
		return e.JSON(http.StatusInternalServerError, msg)
	}
	return e.JSON(http.StatusOK, group)
}

func AddGroup(e echo.Context) error {
	var isExist bool
	group := Groups{}
	err := e.Bind(&group)
	if err != nil {
		LocalLog.Println(err)
	}
	// check duplicate in DB
	err = DB.QueryRow(duplicateGroup, group.Name, group.ParentGroupId).Scan(&isExist)
	if err != nil {
		LocalLog.Println(err)
	}
	if isExist {
		msg.Msg = "Duplicate of the group in the database"
		return e.JSON(http.StatusBadRequest, msg)
	}

	if group.ParentGroupId != 0 {
		// check esist parent_group_id
		err = DB.QueryRow(existGroup, group.ParentGroupId).Scan(&isExist)
		if err != nil {
			LocalLog.Println(err)
		}
		if !isExist {
			msg.Msg = "You specified a non-existent parent"
			return e.JSON(http.StatusBadRequest, msg)
		}
	}
	// add info about new group in db
	err = DB.QueryRow(addGroup, group.Name, group.ParentGroupId).Scan(&group.Id, &group.Name, &group.ParentGroupId)
	if err != nil {
		LocalLog.Println(err)
	}
	return e.JSON(http.StatusCreated, group)
}

type AAA struct {
	Name string
	MyId int
}

type BBB struct {
	AAA
	NewName string
}

func MoveGroup(e echo.Context) error {
	var isExist bool
	var currentGroupId int
	newParentGroup := NewParentGroup{}
	group := Groups{}
	err := e.Bind(&newParentGroup)
	if err != nil {
		LocalLog.Println(err)
	}
	err = DB.QueryRow(
		duplicateGroup, newParentGroup.Name, newParentGroup.NewParentGroupId).Scan(&isExist)
	if err != nil {
		LocalLog.Println(err)
	}
	if isExist {
		msg.Msg = fmt.Sprintf("The group: %d already exists, the sub-group with the same name: %s", newParentGroup.NewParentGroupId, newParentGroup.Name)
		return e.JSON(http.StatusBadRequest, msg)
	}
	err = DB.QueryRow(
		loopGroup, newParentGroup.Name, newParentGroup.ParentGroupId).Scan(&currentGroupId)
	if err != nil {
		LocalLog.Println(err)
	}
	if currentGroupId == newParentGroup.NewParentGroupId {
		msg.Msg = "Not a parent group can specify their group"
		return e.JSON(http.StatusBadRequest, msg)
	}

	if isExist {
		msg.Msg = "You are trying to move the group into yourself. Try again:)"
		return e.JSON(http.StatusBadRequest, msg)
	}
	// add info about new group in db
	err = DB.QueryRow(
		updateParentGroup, newParentGroup.NewParentGroupId,
		newParentGroup.Name, newParentGroup.ParentGroupId).Scan(&group.Id, &group.Name, &group.ParentGroupId)
	switch {
	case err == sql.ErrNoRows:
		return e.JSON(http.StatusNotFound, nil)
	case err != nil:
		LocalLog.Println(err)
		msg.Msg = msgInternalError
		return e.JSON(http.StatusInternalServerError, msg)
	}
	return e.JSON(http.StatusOK, group)
}

func DeleteGroup(e echo.Context) error {
	requestId := RequestId{}
	err := e.Bind(&requestId)
	if err != nil {
		LocalLog.Println(err)
	}
	_, err = DB.Query(deleteTreeGroup, requestId.Id)
	if err != nil {
		LocalLog.Println(err)
		msg.Msg = msgInternalError
		return e.JSON(http.StatusInternalServerError, msg)
	}
	return e.JSON(http.StatusOK, nil)
}

func GetTreeGroup(e echo.Context) error {
	groups := []Groups{}
	infoTree := GetTree{}
	err := e.Bind(&infoTree)
	if err != nil {
		LocalLog.Println(err)
	}
	rows, err := DB.Query(selectTreeGroup, infoTree.Id, infoTree.Depth)
	switch {
	case err == sql.ErrNoRows:
		return e.JSON(http.StatusNotFound, nil)
	case err != nil:
		LocalLog.Println(err)
		msg.Msg = msgInternalError
		return e.JSON(http.StatusInternalServerError, msg)
	}

	for rows.Next() {
		group := Groups{}
		err = rows.Scan(&group.Id, &group.Name, &group.ParentGroupId)
		if err != nil {
			LocalLog.Println(err)
		}
		groups = append(groups, group)
	}
	return e.JSON(http.StatusOK, groups)
}
