package booking

type Service struct {
	store BookingStore
}

func NewService(store BookingStore) *Service {
	return &Service{store: store}
}

func (service *Service) Book(booking Booking) error {
	return service.store.Book(booking)
}
