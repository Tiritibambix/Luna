package handlers

import (
	"luna-backend/api/internal/util"
	"luna-backend/auth"
	"luna-backend/crypto"
	"luna-backend/errors"
	"luna-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error messages are intentionally kept vague in lower verbosity levels,
// because detailed error messages about authenticatino checks might pose a
// security risk.

func Login(c *gin.Context, body *struct {
	Username string `json:"username" form:"username" binding:"alphanumunicode,required"`
	Password string `json:"password" form:"password" binding:"required"`
}) {
	// Parsing
	u := util.GetUtil(c)

	usernameErr := util.IsValidUsername(body.Username)
	passwordErr := util.IsValidPassword(body.Password)
	if usernameErr != nil || passwordErr != nil {
		u.Error(errors.New().Status(http.StatusBadRequest).
			AddErr(errors.LvlDebug, usernameErr).AndErr(passwordErr).
			Append(errors.LvlDebug, "Input did not pass validation").
			Append(errors.LvlWordy, "Malformed request").
			Append(errors.LvlBroad, "Could not log in"),
		)
		return
	}

	// Check if the user exists
	userId, err := u.Tx.Queries().GetUserIdFromUsername(body.Username)
	if err != nil {
		u.Error(err.Status(http.StatusUnauthorized).
			Append(errors.LvlDebug, "Could not find ID for user %v", body.Username).
			Append(errors.LvlPlain, "Invalid credentials").
			Append(errors.LvlBroad, "Could not log in"),
		)

		// Hash the wrong password to prevent timing attacks
		_, _ = auth.SecurePassword(body.Password, u.Config)

		return
	}

	// Get the user's password
	savedPassword, err := u.Tx.Queries().GetPassword(userId)
	if err != nil {
		u.Error(err.Status(http.StatusUnauthorized).
			Append(errors.LvlDebug, "Could not get password for user %v", userId.String()).
			Append(errors.LvlPlain, "Invalid credentials").
			Append(errors.LvlBroad, "Could not log in"),
		)

		// Hash the wrong password to prevent timing attacks
		_, _ = auth.SecurePassword(body.Password, u.Config)

		return
	}

	// Verify the password
	if !auth.VerifyPassword(body.Password, savedPassword, u.Config) {
		u.Error(errors.New().Status(http.StatusUnauthorized).
			Append(errors.LvlDebug, "Wrong password").
			Append(errors.LvlPlain, "Invalid credentials").
			Append(errors.LvlBroad, "Could not log in"),
		)
		return
	}

	// Silently update the user's password to a newer algorithm if applicable
	if !auth.PasswordStillSecure(savedPassword) {
		u.Logger.Infof("updating password %v for user to newer algorithm", body.Username)
		newPassword, err := auth.SecurePassword(body.Password, u.Config)
		if err != nil {
			u.Error(err.
				Append(errors.LvlDebug, "Could not rehash password").
				Append(errors.LvlWordy, "Internal server error").
				Append(errors.LvlBroad, "Could not log in"),
			)
			return
		}
		err = u.Tx.Queries().UpdatePassword(userId, newPassword)
		if err != nil {
			u.Error(err.
				Append(errors.LvlDebug, "Could not update password").
				Append(errors.LvlWordy, "Database error").
				Append(errors.LvlBroad, "Could not log in"),
			)
			return
		}
	}

	// Check if the user account is disabled
	enabled, err := u.Tx.Queries().IsUserEnabled(userId)
	if err != nil {
		u.Error(err.
			Append(errors.LvlDebug, "Could not check if user %v is enabled", userId.String()).
			Append(errors.LvlWordy, "Database error").
			Append(errors.LvlBroad, "Could not log in"),
		)
		return
	}
	if !enabled {
		u.Error(errors.New().Status(http.StatusForbidden).
			Append(errors.LvlPlain, "Your account is disabled."),
		)
		return
	}

	// Create new session
	secret, err := crypto.GenerateRandomBytes(256)
	if err != nil {
		u.Error(err.
			Append(errors.LvlWordy, "Could not generate random bytes").
			AltStr(errors.LvlBroad, "Could not create API key"),
		)
		return
	}

	session := &types.Session{
		UserId:           userId,
		UserAgent:        c.Request.UserAgent(),
		LastIpAddress:    util.DetermineClientAddress(c),
		InitialIpAddress: util.DetermineClientAddress(c),
		IsShortLived:     c.PostForm("remember") != "true",
		IsApi:            false,
		SecretHash:       []byte{},
	}
	err = u.Tx.Queries().InsertSession(session)
	if err != nil {
		u.Error(err.
			Append(errors.LvlBroad, "Could not log in"),
		)
		return
	}

	serverSecret, tr := crypto.GetSymmetricKey(u.Config, "tokenHashSecret")
	if tr != nil {
		u.Error(tr)
		c.Abort()
		return
	}
	err = u.Tx.Queries().UpdateSessionHash(session.SessionId, crypto.GetSha256Hash(serverSecret, session.SessionId.Bytes(), secret))
	if err != nil {
		u.Error(err.
			Append(errors.LvlBroad, "Could not log in"),
		)
		return
	}

	// Generate the token
	token, err := auth.NewToken(u.Config, u.Tx, userId, session.SessionId, secret)
	if err != nil {
		u.Error(err.
			Append(errors.LvlWordy, "Could not generate token").
			Append(errors.LvlBroad, "Could not log in"),
		)
		return
	}

	u.Success(&gin.H{"token": token})
}

func Register(c *gin.Context, body *struct {
	Username   string `json:"username" form:"username" binding:"alphanumunicode,required"`
	Password   string `json:"password" form:"password" binding:"required"`
	Email      string `json:"email" form:"email" binding:"email,required"`
	InviteCode string `json:"invite_code" form:"invite_code"`
}) {
	u := util.GetUtil(c)

	// Check if any users exist to know if this user should be an admin
	usersExist, err := u.Tx.Queries().AnyUsersExist()
	if err != nil {
		u.Error(err.
			Append(errors.LvlDebug, "Could not check if any users exist").
			Append(errors.LvlWordy, "Database error").
			Append(errors.LvlBroad, "Could not register"),
		)
		return
	}

	usernameErr := util.IsValidUsername(body.Username)
	passwordErr := util.IsValidPassword(body.Password)
	if usernameErr != nil || passwordErr != nil {
		u.Error(errors.New().Status(http.StatusBadRequest).
			AddErr(errors.LvlDebug, usernameErr).AndErr(passwordErr).
			Append(errors.LvlDebug, "Input did not pass validation").
			Append(errors.LvlWordy, "Malformed request").
			Append(errors.LvlPlain, "Could not register"),
		)
		return
	}

	// Check invite code and remove it from the database
	var invite *types.RegistrationInvite
	if body.InviteCode != "" {
		invite, err = u.Tx.Queries().GetValidInvite(body.Email, body.InviteCode)
		if err != nil {
			u.Error(err)
			return
		}
		if invite == nil {
			u.Error(errors.New().Status(http.StatusForbidden).
				Append(errors.LvlPlain, "Invalid invite code"),
			)
			return
		}
		u.Tx.Queries().DeleteInvite(invite.InviteId)
	}

	// Check if registration is enabled or the user is the first user
	if !u.Config.Settings.RegistrationEnabled.Enabled && usersExist && invite == nil {
		u.Error(errors.New().Status(http.StatusForbidden).
			Append(errors.LvlWordy, "Open registration is disabled").
			AltStr(errors.LvlPlain, "Registration is disabled").
			Append(errors.LvlBroad, "Could not register"),
		)
		return
	}

	// Hash the password
	securedPassword, err := auth.SecurePassword(body.Password, u.Config)
	if err != nil {
		u.Error(err.
			Append(errors.LvlDebug, "Could not hash password").
			Append(errors.LvlWordy, "Internal server error").
			Append(errors.LvlBroad, "Could not register"),
		)
		return
	}

	// Construct the user
	user := &types.User{
		Username:           body.Username,
		Email:              body.Email,
		Admin:              !usersExist,
		Searchable:         true,
		ProfilePictureType: "static",
		ProfilePictureFile: types.EmptyId(),
		ProfilePictureUrl:  util.GetDefaultProfilePictureUrl(!u.Config.Settings.EnableGravatar.Enabled, body.Email),
	}

	// Insert the user into the database
	userId, err := u.Tx.Queries().AddUser(user)
	if err != nil {
		u.Error(err.
			Append(errors.LvlBroad, "Could not register"),
		)
		return
	}

	// Initialize the user's settings
	err = u.Tx.Queries().InitializeUserSettings(userId)
	if err != nil {
		u.Error(err.
			Append(errors.LvlDebug, "Could not register"),
		)
		return
	}

	// Insert the password into the database
	err = u.Tx.Queries().InsertPassword(user.Id, securedPassword)
	if err != nil {
		u.Error(err.
			Append(errors.LvlDebug, "Could not insert password").
			Append(errors.LvlWordy, "Internal server error").
			Append(errors.LvlBroad, "Could not register"),
		)
		return
	}

	// Create new session
	secret, err := crypto.GenerateRandomBytes(256)
	if err != nil {
		u.Error(err.
			Append(errors.LvlWordy, "Could not generate random bytes").
			AltStr(errors.LvlBroad, "Could not create API key"),
		)
		return
	}

	session := &types.Session{
		UserId:           userId,
		UserAgent:        c.Request.UserAgent(),
		InitialIpAddress: util.DetermineClientAddress(c),
		LastIpAddress:    util.DetermineClientAddress(c),
		IsShortLived:     c.PostForm("remember") != "true",
		IsApi:            false,
		SecretHash:       []byte{},
	}
	err = u.Tx.Queries().InsertSession(session)
	if err != nil {
		u.Error(err.
			Append(errors.LvlBroad, "Could not register"),
		)
		return
	}

	serverSecret, tr := crypto.GetSymmetricKey(u.Config, "tokenHashSecret")
	if tr != nil {
		u.Error(tr)
		c.Abort()
		return
	}
	err = u.Tx.Queries().UpdateSessionHash(session.SessionId, crypto.GetSha256Hash(serverSecret, session.SessionId.Bytes(), secret))
	if err != nil {
		u.Error(err.
			Append(errors.LvlBroad, "Could not register"),
		)
		return
	}

	// Generate the token
	token, err := auth.NewToken(u.Config, u.Tx, userId, session.SessionId, secret)
	if err != nil {
		u.Error(err.
			Append(errors.LvlWordy, "Could not generate token").
			Append(errors.LvlBroad, "Could not register"),
		)
		return
	}

	u.Success(&gin.H{"token": token})
}

func RegistrationEnabled(c *gin.Context) {
	u := util.GetUtil(c)

	usersExist, err := u.Tx.Queries().AnyUsersExist()
	if err != nil {
		u.Error(err.
			Append(errors.LvlDebug, "Could not check if any users exist").
			Append(errors.LvlWordy, "Database error"),
		)
		return
	}

	u.Success(&gin.H{
		"enabled": u.Config.Settings.RegistrationEnabled.Enabled || !usersExist,
	})
}
