package domain

type User struct {
	Id       UserId   `json:"id"`
	Name     string   `json:"name"`
	JoinDate DateTime `json:"joinDate"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
}

type UserId string

func (d *Domain) GetUser(id UserId) (User, error) {
	transId, err := d.Repo.CreateTransaction()
	if err != nil {
		return User{}, err
	}
	u, err := d.Repo.User(id, transId)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (d *Domain) UpdateUser(u User) (User, error) {
	var err error
	if err = d.validateUser(u); err != nil {
		return User{}, err
	}
	transId, err := d.Repo.CreateTransaction()
	defer d.Repo.CommitTransaction(transId)
	if err != nil {
		return User{}, err
	}
	u, err = d.Repo.UpdateUser(u, transId)
	if err != nil {
		return User{}, err
	}
	err = d.Repo.CommitTransaction(transId)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (d *Domain) DeleteUser(id UserId) error {
	transId, err := d.Repo.CreateTransaction()
	if err != nil {
		return err
	}

	// Recursive delete all child resources
	defer d.Repo.RollbackTransaction(transId)
	if _, err = d.Repo.User(id, transId); err != nil {
		return err
	}
	var userTrips []Trip
	userTrips, err = d.Repo.GetUserTrips(id, transId)
	if err != nil {
		return err
	}
	for _, t := range userTrips {
		if err = d.Repo.DeleteTrip(t.Id, transId); err != nil {
			return err
		}
	}
	return d.Repo.CommitTransaction(transId)
}

func (d *Domain) validateUser(u User) error {
	transId, err := d.Repo.CreateTransaction()
	if err != nil {
		return err
	}
	if _, err = d.Repo.User(u.Id, transId); err != nil {
		return err
	}
	return nil
}
