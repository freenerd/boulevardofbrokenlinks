package main

type user struct {
}

type githubUser struct {
  GithubId int
  AccessToken string
  Login string
  Email string
  CheckUrl string
}

func (u *githubUser) Save() error {
  var id int
  err := db.conn.QueryRow(`INSERT INTO users_github(
    github_id, login, access_token, email, check_url
  ) VALUES (
    $1, $2, $3, $4, $5) RETURNING github_id`,
    u.GithubId,
    u.Login,
    u.AccessToken,
    u.Email,
    u.CheckUrl).Scan(&id)
  if err != nil {
    return err
  }
  return nil
}
