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
	transId, err := d.repo.CreateTransaction()
	if err != nil {
		return User{}, err
	}
	u, err := d.repo.User(id, transId)
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
	transId, err := d.repo.CreateTransaction()
	defer d.repo.CommitTransaction(transId)
	if err != nil {
		return User{}, err
	}
	u, err = d.repo.UpdateUser(u, transId)
	if err != nil {
		return User{}, err
	}
	err = d.repo.CommitTransaction(transId)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (d *Domain) DeleteUser(id UserId) error {
	transId, err := d.repo.CreateTransaction()
	if err != nil {
		return err
	}

	// Recursive delete all child resources
	defer d.repo.RollbackTransaction(transId)
	if _, err = d.repo.User(id, transId); err != nil {
		return err
	}
	var userTrips []Trip
	userTrips, err = d.repo.GetUserTrips(id, transId)
	if err != nil {
		return err
	}
	for _, t := range userTrips {
		if err = d.repo.DeleteTrip(t.Id, transId); err != nil {
			return err
		}
	}
	return d.repo.CommitTransaction(transId)
}

func (d *Domain) validateUser(u User) error {
	transId, err := d.repo.CreateTransaction()
	if err != nil {
		return err
	}
	if _, err = d.repo.User(u.Id, transId); err != nil {
		return err
	}
	return nil
}
