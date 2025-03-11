package drawref

import drs3 "github.com/drawref/drawref-backend/drawref/s3"

var TheS3 *drs3.DRS3

func OpenS3(bucket, uploadKeyPrefix, uploadURLPrefix string) error {
	thisS3, err := drs3.Open(bucket, uploadKeyPrefix, uploadURLPrefix)
	if err != nil {
		return err
	}

	TheS3 = thisS3

	return nil
}
