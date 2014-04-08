package main

import (
	"fmt"
)

type user struct {
}

type githubUser struct {
	GithubId    int
	AccessToken string
	Login       string
	Email       string
	CheckUrl    string
}

// This is terrible and should be a proper Upsert instead
// I just hacked this together quickly, please improve!
func (u *githubUser) Save() error {
	var resp int
	err1 := db.conn.QueryRow(`
    UPDATE users_github SET
      login=$2, access_token=$3, email=$4, check_url=$5
    WHERE github_id=$1
    RETURNING github_id;
    `,
		u.GithubId,
		u.Login,
		u.AccessToken,
		u.Email,
		u.CheckUrl,
	).Scan(&resp)

	err2 := db.conn.QueryRow(`
    INSERT INTO users_github(
        github_id, login, access_token, email, check_url
      ) VALUES ($1, $2, $3, $4, $5)
    `,
		u.GithubId,
		u.Login,
		u.AccessToken,
		u.Email,
		u.CheckUrl,
	).Scan(&resp)

	if err1 != nil && err2 != nil {
		return fmt.Errorf("err1: %s err2: %s", err1, err2)
	}
	return nil
}

func GetUser(id int) (*githubUser, error) {
	rows, err := db.conn.Query(`SELECT
      github_id, login, access_token, email, check_url
      FROM users_github
      WHERE github_id=$1
    `,
		id,
	)
	if err != nil {
		return nil, err
	}

	// only do once, since we assume only one id since github_id unique
	rows.Next()
	u := githubUser{}
	err = rows.Scan(
		&u.GithubId,
		&u.Login,
		&u.AccessToken,
		&u.Email,
		&u.CheckUrl,
	)
	if err != nil {
		return nil, err
	}
	if u.GithubId == 0 {
		return nil, fmt.Errorf("User not found:")
	}

	return &u, nil
}
