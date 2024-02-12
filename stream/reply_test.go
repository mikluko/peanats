package stream

import "testing"

func benchmarkReplySubject(b *testing.B, subj ReplySubjecter) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = subj.ReplySubject()
	}
}

func BenchmarkReplySubject(b *testing.B) {
	b.Run("nuid", func(b *testing.B) { benchmarkReplySubject(b, ReplySubjectNUID()) })
	b.Run("math/rand", func(b *testing.B) { benchmarkReplySubject(b, ReplySubjectRand()) })
	b.Run("crypto/rand", func(b *testing.B) { benchmarkReplySubject(b, ReplySubjectCryptoRand()) })
}
