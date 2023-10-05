package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/request"
	log "github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

const (
	outboxPageSize = 50
)

// OutboxHandler will handle requests for the local ActivityPub outbox.
func GetOutboxHandler(
	signer *crypto.Signer,
	actor *models.Actor,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var response interface{}
		var err error
		if r.URL.Query().Get("page") != "" {
			response, err = getOutboxPage(r.URL.Query().Get("page"), r)
		} else {
			response, err = getInitialOutboxHandler(r)
		}

		if response == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		actorIRI := actor.GetActorIri()
		publicKey := crypto.GetPublicKey(actorIRI, actor.PublicKey)
		privateKey := crypto.GetPrivateKey(actor.PrivateKey.String)

		sigResponse := request.NewSignedResponse(signer)

		if err := sigResponse.WriteStreamResponse(response.(vocab.Type), w, publicKey, privateKey); err != nil {
			log.Errorln("unable to write stream response for outbox handler", err)
		}
	}
}

func getInitialOutboxHandler(r *http.Request) (vocab.ActivityStreamsOrderedCollection, error) {
	collection := streams.NewActivityStreamsOrderedCollection()

	//idProperty := streams.NewJSONLDIdProperty()
	//id, err := createPageURL(r, nil)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to create followers page property")
	//}
	//idProperty.SetIRI(id)
	//collection.SetJSONLDId(idProperty)
	//
	//totalPosts, err := persistence.GetOutboxPostCount()
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to get outbox post count")
	//}
	//totalItemsProperty := streams.NewActivityStreamsTotalItemsProperty()
	//totalItemsProperty.Set(int(totalPosts))
	//collection.SetActivityStreamsTotalItems(totalItemsProperty)
	//
	//first := streams.NewActivityStreamsFirstProperty()
	//page := "1"
	//firstIRI, err := createPageURL(r, &page)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to create first page property")
	//}
	//
	//first.SetIRI(firstIRI)
	//collection.SetActivityStreamsFirst(first)

	return collection, nil
}

func getOutboxPage(page string, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	//pageInt, err := strconv.Atoi(page)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to parse page number")
	//}
	//
	//postCount, err := persistence.GetOutboxPostCount()
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to get outbox post count")
	//}
	//
	collectionPage := streams.NewActivityStreamsOrderedCollectionPage()
	//idProperty := streams.NewJSONLDIdProperty()
	//id, err := createPageURL(r, &page)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to create followers page ID")
	//}
	//idProperty.SetIRI(id)
	//collectionPage.SetJSONLDId(idProperty)
	//
	//orderedItems := streams.NewActivityStreamsOrderedItemsProperty()
	//
	//outboxItems, err := persistence.GetOutbox(outboxPageSize, (pageInt-1)*outboxPageSize)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to get federation followers")
	//}
	//orderedItems.AppendActivityStreamsOrderedCollection(outboxItems)
	//collectionPage.SetActivityStreamsOrderedItems(orderedItems)
	//
	//partOf := streams.NewActivityStreamsPartOfProperty()
	//partOfIRI, err := createPageURL(r, nil)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to create partOf property for outbox page")
	//}
	//
	//partOf.SetIRI(partOfIRI)
	//collectionPage.SetActivityStreamsPartOf(partOf)
	//
	//if pageInt*followersPageSize < int(postCount) {
	//	next := streams.NewActivityStreamsNextProperty()
	//	nextPage := fmt.Sprintf("%d", pageInt+1)
	//	nextIRI, err := createPageURL(r, &nextPage)
	//	if err != nil {
	//		return nil, errors.Wrap(err, "unable to create next page property")
	//	}
	//
	//	next.SetIRI(nextIRI)
	//	collectionPage.SetActivityStreamsNext(next)
	//}

	return collectionPage, nil
}
