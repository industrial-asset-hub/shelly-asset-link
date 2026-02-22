package scanner

type Publisher interface {
    PublishDiscovery(asset any) error
}
