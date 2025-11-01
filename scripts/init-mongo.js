// Initialize MongoDB with user and database
db = db.getSiblingDB('es_demo');

// Create application user
db.createUser({
  user: 'appuser',
  pwd: 'apppassword',
  roles: [
    {
      role: 'readWrite',
      db: 'es_demo'
    }
  ]
});

// Create collections with indexes
db.createCollection('bank_accounts');

// Add indexes for better query performance
db.bank_accounts.createIndex({ "aggregate_id": 1 }, { unique: true });
db.bank_accounts.createIndex({ "account_number": 1 });
db.bank_accounts.createIndex({ "created_at": -1 });

print('MongoDB initialized successfully');
