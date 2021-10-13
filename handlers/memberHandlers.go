package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs401/letsgo/models"
)

// MemberHandler handler func receiver
type MemberHandler struct{}

// RequestMembershipHandler inserts a pending membership request record.
func (handler *MemberHandler) RequestMembershipHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "thread.gotmpl", gin.H{
			"error": "Badrequest",
			"user":  email,
		})
		return
	}
	db := models.DBConn
	user := getUserByEmail(email)
	var prePending models.PendingMember
	db.Where("forum_id = ?", id).Where("user_id = ?", user.ID).Find(&prePending)
	// If there is no previous pending request
	if prePending.ID == 0 {
		pending := models.PendingMember{
			UserID:  user.ID,
			ForumID: uint(id),
		}
		result := db.Create(&pending)
		if result.Error != nil {
			log.Printf("Error: %v", result.Error)
		}
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/forums/%v", id))
}

// ManageMembersHandler populates and returns member management template
func (handler *MemberHandler) ManageMembersHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "forum.gotmpl", gin.H{
			"error": "Badrequest",
			"user":  email,
		})
		return
	}
	db := models.DBConn

	var members []models.Member
	var pendingMembers []models.PendingMember
	db.Where("forum_id = ?", id).Preload("User").Find(&members)
	db.Where("forum_id = ?", id).Preload("User").Find(&pendingMembers)
	c.HTML(http.StatusOK, "manage_members.gotmpl", gin.H{
		"fid":             id,
		"user":            email,
		"members":         members,
		"pending_members": pendingMembers,
	})
}

// AddMemberHandler adds a user to the member list of the forum
func (handler *MemberHandler) AddMemberHandler(c *gin.Context) {
	db := models.DBConn
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	uid, err := strconv.Atoi(c.Request.FormValue("uid"))
	if err != nil {
		log.Printf("Unable to convert uid form value to int: %v", err)
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Unable to convert id param to int: %v", err)
		session.AddFlash("GroupID invalid")
	}
	var pMember models.PendingMember
	result := db.Find(&pMember, uid)
	if result.Error != nil {
		// Error retrieving the pending request
		log.Printf("Error retrieving the pending request: %v", result.Error.Error())
		c.Redirect(http.StatusFound, fmt.Sprintf("/manage_members/%v", id))
		session.Save()
		return
	}
	if pMember.ID == 0 {
		// pending request doesn't exist
		session.AddFlash("Pending request doesn't exist.")
		c.Redirect(http.StatusFound, fmt.Sprintf("/manage_members/%v", id))
		session.Save()
		return
	}
	user := getUserByEmail(email)
	var forum models.Forum
	db.Find(&forum, id)
	if user != nil && user.ID == forum.UserID {
		var member models.Member
		member.ForumID = pMember.ForumID
		member.UserID = pMember.UserID
		db.Create(&member)
		db.Delete(&pMember)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/manage_members/%v", id))
	session.Save()
}

// RejectMemberHandler removes the pending request
func (handler *MemberHandler) RejectMemberHandler(c *gin.Context) {
	db := models.DBConn
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	uid, err := strconv.Atoi(c.Request.FormValue("uid"))
	if err != nil {
		log.Printf("Unable to convert uid form value to int: %v", err)
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Unable to convert id param to int: %v", err)
		session.AddFlash("GroupID invalid")
	}

	var pMember models.PendingMember
	result := db.Find(&pMember, uid)
	if result.Error != nil {
		// Error retrieving the pending request
		log.Printf("Error retrieving the pending request: %v", result.Error.Error())
		c.Redirect(http.StatusFound, fmt.Sprintf("/manage_members/%v", id))
		session.Save()
		return
	}
	if pMember.ID == 0 {
		// pending request doesn't exist
		session.AddFlash("Pending request doesn't exist.")
		c.Redirect(http.StatusFound, fmt.Sprintf("/manage_members/%v", id))
		session.Save()
		return
	}
	user := getUserByEmail(email)
	var forum models.Forum
	db.Find(&forum, id)
	if user != nil && user.ID == forum.UserID {
		db.Delete(&pMember)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/manage_members/%v", id))
	session.Save()
}

// RemoveMemberHandler removes a member from the member list
func (handler *MemberHandler) RemoveMemberHandler(c *gin.Context) {
	db := models.DBConn
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	uid, err := strconv.Atoi(c.Request.FormValue("uid"))
	if err != nil {
		log.Printf("Unable to convert uid form value to int: %v", err)
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Unable to convert id param to int: %v", err)
		session.AddFlash("GroupID invalid")
	}

	var member models.Member
	result := db.Find(&member, uid)
	if result.Error != nil {
		// Error retrieving the pending request
		log.Printf("Error retrieving member: %v", result.Error.Error())
		c.Redirect(http.StatusFound, fmt.Sprintf("/manage_members/%v", id))
		session.Save()
		return
	}
	if member.ID == 0 {
		// pending request doesn't exist
		session.AddFlash("Member doesn't exist.")
		c.Redirect(http.StatusFound, fmt.Sprintf("/manage_members/%v", id))
		session.Save()
		return
	}
	user := getUserByEmail(email)
	var forum models.Forum
	db.Find(&forum, id)
	if user != nil && user.ID == forum.UserID {
		db.Delete(&member)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/manage_members/%v", id))
	session.Save()
}
