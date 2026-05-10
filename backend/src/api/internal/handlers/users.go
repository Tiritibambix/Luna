package handlers

import (
	"luna-backend/api/internal/util"
	"luna-backend/auth"
	"luna-backend/constants"
	"luna-backend/errors"
	"luna-backend/files"
	"luna-backend/types"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	u := util.GetUtil(c)

	userId := util.GetUserId(c)
	user, err := u.Tx.Queries().GetUser(userId)
	if err != nil {
		u.Error(err)
		return
	}
	err = user.UpdateEffectiveProfilePicture(u.Config.Settings.CacheProfilePictures.Enabled)
	if err != nil {
		u.Error(err)
		return
	}

	u.Success(&gin.H{"user": user})
}

func GetUsers(c *gin.Context) {
	u := util.GetUtil(c)

	userId := util.GetUserId(c)

	all := c.Query("all") == "true"
	if all && !util.HasAdminPrivilegesAndReportError(c) {
		return
	}

	users, tr := u.Tx.Queries().GetUsers(all)
	if tr != nil {
		u.Error(tr)
		return
	}

	for _, user := range users {
		tr = user.UpdateEffectiveProfilePicture(u.Config.Settings.CacheProfilePictures.Enabled)
		if tr != nil {
			u.Error(tr)
			return
		}
	}

	if all {
		u.Success(&gin.H{
			"users":   users,
			"current": userId,
		})
	} else {
		// When all is not set, i.e., the request comes from a non-admin user,
		// cast User to StrippedUser to remove (potentially) sensitive information.
		strippedUsers := make([]*types.StrippedUser, 0, len(users))
		for _, user := range users {
			strippedUser := &types.StrippedUser{
				Id:                         user.Id,
				Username:                   user.Username,
				Admin:                      user.Admin,
				EffectiveProfilePictureUrl: user.EffectiveProfilePictureUrl,
			}
			strippedUsers = append(strippedUsers, strippedUser)
		}
		u.Success(&gin.H{
			"users":   strippedUsers,
			"current": userId,
		})
	}
}

func PatchUserData(c *gin.Context, body *struct {
	NewUsername           *string               `form:"username" json:"username" binding:"omitempty,alphanumunicode"`
	NewEmail              *string               `form:"email" json:"email" binding:"omitempty,email"`
	OldPassword           *string               `form:"password" json:"password"`
	NewPassword           *string               `form:"new_password" json:"new_password"`
	NewProfilePictureType *string               `form:"pfp_type" json:"pfp_type"`
	NewProfilePictureUrl  *types.Url            `form:"pfp_url" json:"pfp_url" binding:"required_if=pfp_type remote"`
	NewProfilePictureFile *multipart.FileHeader `form:"pfp_file" binding:"required_if=pfp_type database"`
	NewSearchable         *bool                 `form:"searchable" json:"searchable"`
}) {
	u := util.GetUtil(c)

	userId := util.GetUserId(c)

	newProfilePictureRequested := body.NewProfilePictureType != nil || body.NewProfilePictureUrl != nil || body.NewProfilePictureFile != nil

	if newProfilePictureRequested {
		if body.NewProfilePictureType == nil {
			u.Error(errors.New().Status(http.StatusBadRequest).
				Append(errors.LvlPlain, "Missing profile picture type"))
			return
		}
		switch *body.NewProfilePictureType {
		case constants.ProfilePictureGravatar:
			if !u.Config.Settings.EnableGravatar.Enabled {
				u.Error(errors.New().Status(http.StatusBadRequest).
					Append(errors.LvlPlain, "Gravatar profile pictures are disabled"))
				return
			}
		case constants.ProfilePictureDatabase:
			if !u.Config.Settings.EnableProfilePicturesUpload.Enabled {
				u.Error(errors.New().Status(http.StatusBadRequest).
					Append(errors.LvlPlain, "Profile picture uploads are disabled"))
				return
			}
		}
	}

	if body.NewUsername == nil && body.NewEmail == nil && body.NewPassword == nil && !newProfilePictureRequested && body.NewSearchable == nil {
		u.Error(errors.New().Status(http.StatusBadRequest).
			Append(errors.LvlPlain, "Nothing to change"))
		return
	}

	if body.NewUsername != nil && util.IsValidUsername(*body.NewUsername) != nil {
		u.Error(errors.New().Status(http.StatusBadRequest).
			Append(errors.LvlPlain, "Invalid username"))
		return
	}
	if body.NewPassword != nil && util.IsValidPassword(*body.NewPassword) != nil {
		u.Error(errors.New().Status(http.StatusBadRequest).
			Append(errors.LvlPlain, "Invalid password"))
		return
	}

	oldUserStruct, tr := u.Tx.Queries().GetUser(userId)
	if tr != nil {
		u.Error(tr)
		return
	}

	// Set new profile picture
	newProfilePictureId := types.EmptyId()
	if newProfilePictureRequested {
		// Delete old profile picture if applicable
		oldPfpFile := *files.GetDatabaseFile(oldUserStruct.ProfilePictureFile)
		if oldPfpFile.GetId() != types.EmptyId() {
			tr = u.Tx.Queries().DeleteFilecache(&oldPfpFile, userId)
			if tr != nil {
				u.Error(tr.
					Append(errors.LvlDebug, "Could not delete old profile picture file %v", oldPfpFile.GetId()).
					Append(errors.LvlPlain, "Could not update profile picture"))
				return
			}
		}

		// Refresh profile picture if the email changes and gravatar is used
		if !newProfilePictureRequested && body.NewEmail != nil && oldUserStruct.ProfilePictureType == "gravatar" {
			newProfilePictureType := "gravatar"
			body.NewProfilePictureType = &newProfilePictureType

			currentGravatarUrl, err := types.NewUrl(oldUserStruct.ProfilePictureUrl.String())
			if err != nil {
				u.Error(errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlWordy, "Could not parse old gravatar profile picture").
					Append(errors.LvlPlain, "Could not update profile picture"),
				)
				return
			}

			body.NewProfilePictureUrl = util.GetGravatarUrlWithParams(*body.NewEmail, currentGravatarUrl.URL().RawQuery)
		}

		// Parse new profile picture
		if newProfilePictureRequested {
			switch *body.NewProfilePictureType {
			case constants.ProfilePictureGravatar:
				fallthrough
			case constants.ProfilePictureStatic:
				fallthrough
			case constants.ProfilePictureRemote:
				if body.NewProfilePictureUrl == nil {
					u.Error(errors.New().Status(http.StatusBadRequest).
						Append(errors.LvlPlain, "Missing profile picture URL"))
					return
				}

			case constants.ProfilePictureDatabase:
				if body.NewProfilePictureFile == nil {
					u.Error(errors.New().Status(http.StatusBadRequest).
						Append(errors.LvlPlain, "Missing profile picture file"))
					return
				}

				pfpFile, err := body.NewProfilePictureFile.Open()
				if err != nil {
					u.Error(errors.New().Status(http.StatusBadRequest).
						AddErr(errors.LvlDebug, err).
						Append(errors.LvlPlain, "Could not open profile picture file"))
					return
				}

				uploadedFile, tr := files.NewDatabaseFileFromContent((*body.NewProfilePictureFile).Filename, pfpFile, userId, u.Tx.Queries())
				if tr != nil {
					u.Error(tr.
						Append(errors.LvlDebug, "Could not create file from content").
						Append(errors.LvlPlain, "Could not upload profile picture"))
					return
				}

				newProfilePictureId = uploadedFile.GetId()

			default:
				u.Error(errors.New().Status(http.StatusBadRequest).
					Append(errors.LvlPlain, "Invalid profile picture type"))
				return
			}

			// Create local profile picture cache
			if *body.NewProfilePictureType == constants.ProfilePictureGravatar || *body.NewProfilePictureType == constants.ProfilePictureRemote {
				pfpFile, tr := files.NewRemoteFile(body.NewProfilePictureUrl, "image/*", auth.NewNoAuth(), userId, u.Tx.Queries())
				if tr != nil {
					u.Error(tr.
						Append(errors.LvlDebug, "Could not create remote file for profile picture").
						Append(errors.LvlPlain, "Could not upload profile picture"),
					)
					return
				}

				newProfilePictureId = pfpFile.GetId()
			}
		}
	}

	// Reauthenticate if needed
	reauthenticationRequired := body.NewUsername != nil || body.NewEmail != nil || body.NewPassword != nil

	if reauthenticationRequired {
		if body.OldPassword == nil {
			u.Error(errors.New().Status(http.StatusBadRequest).
				Append(errors.LvlPlain, "Missing password"))
			return
		}

		// Get the user's password
		savedPassword, err := u.Tx.Queries().GetPassword(userId)
		if err != nil {
			u.Error(err.Status(http.StatusUnauthorized).
				Append(errors.LvlDebug, "Could not get password for user %v", userId.String()).
				Append(errors.LvlPlain, "Invalid credentials"),
			)
			return
		}

		// Verify the password
		if !auth.VerifyPassword(*body.OldPassword, savedPassword, u.Config) {
			u.Error(errors.New().Status(http.StatusUnauthorized).
				Append(errors.LvlDebug, "Wrong password").
				Append(errors.LvlPlain, "Invalid credentials"),
			)
			return
		}
	}

	// Update the user
	var newUserStruct *types.User
	if body.NewUsername != nil || body.NewEmail != nil || body.NewSearchable != nil || newProfilePictureRequested {
		newUserStruct = &types.User{
			Id:                 userId,
			Username:           oldUserStruct.Username,
			Email:              oldUserStruct.Email,
			Searchable:         oldUserStruct.Searchable,
			ProfilePictureType: oldUserStruct.ProfilePictureType,
			ProfilePictureUrl:  oldUserStruct.ProfilePictureUrl,
			ProfilePictureFile: oldUserStruct.ProfilePictureFile,
			Admin:              oldUserStruct.Admin,
		}

		if body.NewUsername != nil {
			newUserStruct.Username = *body.NewUsername
		}
		if body.NewEmail != nil {
			newUserStruct.Email = *body.NewEmail
		}
		if body.NewSearchable != nil {
			newUserStruct.Searchable = *body.NewSearchable
		}
		if newProfilePictureRequested {
			newUserStruct.ProfilePictureType = *body.NewProfilePictureType
			newUserStruct.ProfilePictureUrl = body.NewProfilePictureUrl
			newUserStruct.ProfilePictureFile = newProfilePictureId
		}

		tr = u.Tx.Queries().UpdateUserData(newUserStruct)
		if tr != nil {
			u.Error(tr)
			return
		}
	}

	// Update the password
	if body.NewPassword != nil {
		securedPassword, err := auth.SecurePassword(*body.NewPassword, u.Config)
		if err != nil {
			u.Error(err.
				Append(errors.LvlDebug, "Could not hash new password"),
			)
			return
		}
		err = u.Tx.Queries().UpdatePassword(userId, securedPassword)
		if err != nil {
			u.Error(err.
				Append(errors.LvlDebug, "Could not update password"),
			)
			return
		}
	}

	response := &gin.H{
		"status": "ok",
	}

	if body.NewProfilePictureType != nil {
		tr = newUserStruct.UpdateEffectiveProfilePicture(u.Config.Settings.CacheProfilePictures.Enabled)
		if tr != nil {
			u.Error(tr)
			return
		}
		(*response)["profile_picture"] = newUserStruct.EffectiveProfilePictureUrl.String()
	}

	u.Success(response)
}

func DeleteUser(c *gin.Context) {
	u := util.GetUtil(c)

	executingUserId := util.GetUserId(c)
	affectedUserId, tr := util.GetIdOrDefault(c, "user", "self", executingUserId)
	if tr != nil {
		u.Error(tr)
		return
	}
	if affectedUserId != executingUserId && !util.HasAdminPrivilegesAndReportError(c) {
		return
	}

	// Get the user's password
	password := c.PostForm("password")
	if password == "" {
		u.Error(errors.New().Status(http.StatusBadRequest).
			Append(errors.LvlPlain, "Missing password"),
		)
		return
	}

	// Get the user's password
	savedPassword, err := u.Tx.Queries().GetPassword(executingUserId)
	if err != nil {
		u.Error(err.Status(http.StatusUnauthorized).
			Append(errors.LvlDebug, "Could not get password for user %v", executingUserId.String()).
			Append(errors.LvlPlain, "Invalid credentials"),
		)
		return
	}

	// Verify the password
	if !auth.VerifyPassword(password, savedPassword, u.Config) {
		u.Error(errors.New().Status(http.StatusUnauthorized).
			Append(errors.LvlDebug, "Wrong password").
			Append(errors.LvlPlain, "Invalid credentials"),
		)
		return
	}

	err = u.Tx.Queries().DeleteUser(affectedUserId)
	if err != nil {
		u.Error(err)
		return
	}

	u.Success(nil)
}

func EnableUser(c *gin.Context) {
	u := util.GetUtil(c)

	userId, tr := util.GetId(c, "user")
	if tr != nil {
		u.Error(tr)
		return
	}

	tr = u.Tx.Queries().SetUserEnabled(userId, true)

	if tr != nil {
		u.Error(tr)
		return
	}

	u.Success(nil)
}

func DisableUser(c *gin.Context) {
	u := util.GetUtil(c)

	userId, tr := util.GetId(c, "user")
	if tr != nil {
		u.Error(tr)
		return
	}

	isAdmin, tr := u.Tx.Queries().IsAdmin(userId)
	if tr != nil {
		u.Error(tr)
		return
	}
	if isAdmin {
		u.Error(errors.New().Status(http.StatusBadRequest).
			Append(errors.LvlPlain, "Admin accounts cannot be disabled."),
		)
		return
	}

	tr = u.Tx.Queries().SetUserEnabled(userId, false)

	if tr != nil {
		u.Error(tr)
		return
	}

	tr = u.Tx.Queries().DeleteSessions(userId)
	if tr != nil {
		u.Error(tr)
		return
	}

	u.Success(nil)
}
