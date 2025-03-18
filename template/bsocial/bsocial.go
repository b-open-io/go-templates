package bsocial

import (
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	"github.com/bsv-blockchain/go-templates/template/bitcom"
)

const (
	// Protocol prefixes
	// bitcom.MapPrefix = "1PuQa7K62MiKCtssSLKy1kh56WWU7MtUR5"
	AIPPrefix = "15PciHG22SNLQJXMoSUaWVi7WSqc7hCfva"
	AppName   = "bsocial"
)

// Media types
type MediaType string

const (
	MediaTypeTextPlain    MediaType = "text/plain"
	MediaTypeTextMarkdown MediaType = "text/markdown"
	MediaTypeTextHTML     MediaType = "text/html"
	MediaTypeImagePNG     MediaType = "image/png"
	MediaTypeImageJPEG    MediaType = "image/jpeg"
)

// Encoding types
type Encoding string

const (
	EncodingUTF8   Encoding = "utf-8"
	EncodingBase64 Encoding = "base64"
	EncodingHex    Encoding = "hex"
)

// Context types
type Context string

const (
	ContextTx       Context = "tx"
	ContextChannel  Context = "channel"
	ContextBapID    Context = "bapID"
	ContextProvider Context = "provider"
	ContextVideoID  Context = "videoID"
)

// Post represents a new piece of content
type Post struct {
	MediaType       MediaType `json:"mediaType"`
	Encoding        Encoding  `json:"encoding"`
	Content         string    `json:"content"`
	Context         Context   `json:"context,omitempty"`
	ContextValue    string    `json:"contextValue,omitempty"`
	Subcontext      Context   `json:"subcontext,omitempty"`
	SubcontextValue string    `json:"subcontextValue,omitempty"`
	Tags            []string  `json:"tags,omitempty"`
	Attachments     []B       `json:"attachments,omitempty"`
}

// B represents B protocol data
type B struct {
	MediaType MediaType `json:"mediaType"`
	Encoding  Encoding  `json:"encoding"`
	Data      Data      `json:"data"`
}

// Data represents the actual content in B protocol
type Data struct {
	UTF8   string `json:"utf8,omitempty"`
	Base64 string `json:"base64,omitempty"`
	Hex    string `json:"hex,omitempty"`
}

// CreatePost creates a new post transaction
func CreatePost(post Post, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte("B"))
	s.AppendPushData([]byte(post.Content))
	s.AppendPushData([]byte(string(post.MediaType)))
	s.AppendPushData([]byte(string(post.Encoding)))
	s.AppendPushData([]byte("UTF8"))

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Create MAP protocol output
	mapScript := &script.Script{}
	mapScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	mapScript.AppendPushData([]byte(bitcom.MapPrefix))
	mapScript.AppendPushData([]byte("SET"))
	mapScript.AppendPushData([]byte("app"))
	mapScript.AppendPushData([]byte(AppName))
	mapScript.AppendPushData([]byte("type"))
	mapScript.AppendPushData([]byte("post"))

	// Add context if provided
	if post.Context != "" {
		mapScript.AppendPushData([]byte("context_" + string(post.Context)))
		mapScript.AppendPushData([]byte(post.ContextValue))
	}

	// Add subcontext if provided
	if post.Subcontext != "" {
		mapScript.AppendPushData([]byte("subcontext_" + string(post.Subcontext)))
		mapScript.AppendPushData([]byte(post.SubcontextValue))
	}

	// Add AIP signature
	mapScript.AppendPushData([]byte("|"))
	mapScript.AppendPushData([]byte(AIPPrefix))
	mapScript.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	mapScript.AppendPushData(pubKey.Compressed())

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: mapScript,
		Satoshis:      0,
	})

	// Add tags if present
	if len(post.Tags) > 0 {
		tagsScript := &script.Script{}
		tagsScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
		tagsScript.AppendPushData([]byte(bitcom.MapPrefix))
		tagsScript.AppendPushData([]byte("SET"))
		tagsScript.AppendPushData([]byte("app"))
		tagsScript.AppendPushData([]byte(AppName))
		tagsScript.AppendPushData([]byte("type"))
		tagsScript.AppendPushData([]byte("post"))
		tagsScript.AppendPushData([]byte("tags"))
		for _, tag := range post.Tags {
			tagsScript.AppendPushData([]byte(tag))
		}
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: tagsScript,
			Satoshis:      0,
		})
	}

	return tx, nil
}

// CreateReply creates a reply to an existing post
func CreateReply(reply Post, replyTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte("B"))
	s.AppendPushData([]byte(reply.Content))
	s.AppendPushData([]byte(string(reply.MediaType)))
	s.AppendPushData([]byte(string(reply.Encoding)))
	s.AppendPushData([]byte("UTF8"))

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Create MAP protocol output
	mapScript := &script.Script{}
	mapScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	mapScript.AppendPushData([]byte(bitcom.MapPrefix))
	mapScript.AppendPushData([]byte("SET"))
	mapScript.AppendPushData([]byte("app"))
	mapScript.AppendPushData([]byte(AppName))
	mapScript.AppendPushData([]byte("type"))
	mapScript.AppendPushData([]byte("post"))
	mapScript.AppendPushData([]byte("context_tx"))
	mapScript.AppendPushData([]byte(replyTxID))

	// Add AIP signature
	mapScript.AppendPushData([]byte("|"))
	mapScript.AppendPushData([]byte(AIPPrefix))
	mapScript.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	mapScript.AppendPushData(pubKey.Compressed())

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: mapScript,
		Satoshis:      0,
	})

	return tx, nil
}

// CreateLike creates a like transaction
func CreateLike(likeTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte("like"))
	s.AppendPushData([]byte("tx"))
	s.AppendPushData([]byte(likeTxID))
	s.AppendPushData([]byte("|"))
	s.AppendPushData([]byte(AIPPrefix))
	s.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	s.AppendPushData(pubKey.Compressed())

	// TODO: Add proper signature calculation

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// TODO: Add proper UTXO handling and change address

	return tx, nil
}

// CreateUnlike creates an unlike transaction
func CreateUnlike(unlikeTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte("unlike"))
	s.AppendPushData([]byte("tx"))
	s.AppendPushData([]byte(unlikeTxID))
	s.AppendPushData([]byte("|"))
	s.AppendPushData([]byte(AIPPrefix))
	s.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	s.AppendPushData(pubKey.Compressed())

	// TODO: Add proper signature calculation

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// TODO: Add proper UTXO handling and change address

	return tx, nil
}

// CreateFollow creates a follow transaction
func CreateFollow(followBapID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte("follow"))
	s.AppendPushData([]byte("bapID"))
	s.AppendPushData([]byte(followBapID))
	s.AppendPushData([]byte("|"))
	s.AppendPushData([]byte(AIPPrefix))
	s.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	s.AppendPushData(pubKey.Compressed())

	// TODO: Add proper signature calculation

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// TODO: Add proper UTXO handling and change address

	return tx, nil
}

// CreateUnfollow creates an unfollow transaction
func CreateUnfollow(unfollowBapID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte("unfollow"))
	s.AppendPushData([]byte("bapID"))
	s.AppendPushData([]byte(unfollowBapID))
	s.AppendPushData([]byte("|"))
	s.AppendPushData([]byte(AIPPrefix))
	s.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	s.AppendPushData(pubKey.Compressed())

	// TODO: Add proper signature calculation

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// TODO: Add proper UTXO handling and change address

	return tx, nil
}

// Message represents a message in a channel or to a user
type Message struct {
	MediaType    MediaType `json:"mediaType"`
	Encoding     Encoding  `json:"encoding"`
	Content      string    `json:"content"`
	Context      Context   `json:"context"`
	ContextValue string    `json:"contextValue"`
}

// CreateMessage creates a new message transaction
func CreateMessage(msg Message, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte("B"))
	s.AppendPushData([]byte(msg.Content))
	s.AppendPushData([]byte(string(msg.MediaType)))
	s.AppendPushData([]byte(string(msg.Encoding)))
	s.AppendPushData([]byte("UTF8"))

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Create MAP protocol output
	mapScript := &script.Script{}
	mapScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	mapScript.AppendPushData([]byte(bitcom.MapPrefix))
	mapScript.AppendPushData([]byte("SET"))
	mapScript.AppendPushData([]byte("app"))
	mapScript.AppendPushData([]byte(AppName))
	mapScript.AppendPushData([]byte("type"))
	mapScript.AppendPushData([]byte("message"))
	mapScript.AppendPushData([]byte("context_" + string(msg.Context)))
	mapScript.AppendPushData([]byte(msg.ContextValue))

	// Add AIP signature
	mapScript.AppendPushData([]byte("|"))
	mapScript.AppendPushData([]byte(AIPPrefix))
	mapScript.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	mapScript.AppendPushData(pubKey.Compressed())

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: mapScript,
		Satoshis:      0,
	})

	return tx, nil
}
