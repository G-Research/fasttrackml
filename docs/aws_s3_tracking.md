# FastTrackML with AWS S3 Setup

To use FastTrackML with AWS S3 for artifact storage, you'll need to configure your environment variables and make sure the AWS S3 bucket is set up correctly.

## Environment Variables

`AWS_ACCESS_KEY_ID`
Your AWS access key ID in the format `AKIAxxxxxxxxxxxx` (20 characters).

`AWS_SECRET_ACCESS_KEY`
Your AWS secret access key (40 characters).

`AWS_DEFAULT_REGION`
The AWS region where your S3 bucket is located. Example: `us-east-1`, `us-west-2`, etc.

`FML_S3_ENDPOINT_URI`
The endpoint URI for your S3 bucket. This will change depending on the path-style access type. If path-style access is disabled, exclude the bucket name subdomain. Example: `https://s3.your-region.amazonaws.com`. Otherwise, the bucket subdomain can be included: `https://your-bucket-name.s3.your-region.amazonaws.com`.

<!-- FML_S3_USE_PATH_STYLE
    Description: Determines whether to use path-style access when interacting with S3.
    Format: true or false
    Notes: This should typically be set to true for older S3 buckets or S3-compatible storage that requires path-style access. -->

## AWS S3 Bucket Setup

### Bucket Permissions:
Ensure your bucket has the necessary permissions for the access keys you're using. It must have `s3:GetObject`, `s3:PutObject`, and `s3:DeleteObject` permissions.

If you're encountering "Access Denied" errors, check the bucket's policy to grant appropriate access to the user/role associated with your access keys.

## Troubleshooting
The following errors may show up in the frontend if there are issues with your AWS S3 setup:

### 403 Forbidden:
1. Double-check your `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.
2. Check that the bucket policy allows access from the IP range or VPC you are working from.
3. If using a non-AWS S3-compatible storage such as Minio, verify that `FML_S3_ENDPOINT_URI` is correct.

### 404 (Object Does Not Exist)
1. Check that the object key format matches the actual key in your S3 bucket.
<!-- 2. Make sure that `FML_S3_USE_PATH_STYLE` is set correctly based on your bucket's URL structure. -->