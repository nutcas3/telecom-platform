#!/bin/bash
set -e

echo "Setting up MongoDB for TaaS Platform..."

# Check if MongoDB is installed
if ! command -v mongosh &> /dev/null; then
    echo "MongoDB shell (mongosh) not found. Please install MongoDB first."
    echo "Visit: https://www.mongodb.com/docs/manual/installation/"
    exit 1
fi

# MongoDB connection details
MONGO_HOST=${MONGO_HOST:-"localhost"}
MONGO_PORT=${MONGO_PORT:-"27017"}
MONGO_USER=${MONGO_USER:-"free5gc"}
MONGO_PASS=${MONGO_PASS:-"free5gc_password"}
MONGO_DB="free5gc"

echo "Creating MongoDB database and user..."

# Create database and user
mongosh --host $MONGO_HOST --port $MONGO_PORT <<EOF
use admin

// Create user if not exists
if (!db.getUser("$MONGO_USER")) {
    db.createUser({
        user: "$MONGO_USER",
        pwd: "$MONGO_PASS",
        roles: [
            { role: "readWrite", db: "$MONGO_DB" },
            { role: "dbAdmin", db: "$MONGO_DB" }
        ]
    })
    print("User $MONGO_USER created successfully")
} else {
    print("User $MONGO_USER already exists")
}

use $MONGO_DB

// Create collections for free5GC
db.createCollection("subscriptionData.provisionedData.amData")
db.createCollection("subscriptionData.authenticationData.authenticationSubscription")
db.createCollection("subscriptionData.provisionedData.smData")
db.createCollection("policyData.ues.amData")
db.createCollection("policyData.ues.smData")

print("Collections created successfully")
EOF

echo "MongoDB setup complete!"
echo ""
echo "Connection string: mongodb://$MONGO_USER:$MONGO_PASS@$MONGO_HOST:$MONGO_PORT/$MONGO_DB"
echo ""
echo "Update your .env files with:"
echo "MONGODB_URI=mongodb://$MONGO_USER:$MONGO_PASS@$MONGO_HOST:$MONGO_PORT/$MONGO_DB"
