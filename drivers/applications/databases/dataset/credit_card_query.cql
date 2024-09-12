// Load the CSV file and create nodes and relationships
LOAD CSV WITH HEADERS FROM 'file:///fraud_data.csv' AS row

// Create or merge Merchant nodes
MERGE (m:Merchant {name: row.merchant})
ON CREATE SET m.latitude = toFloat(row.merch_lat),
              m.longitude = toFloat(row.merch_long)

// Create Transaction nodes
CREATE (t:Transaction {
    transactionId: row.trans_num,
    transDateTime: row.trans_date_trans_time,
    amount: toFloat(row.amt),
    city: row.city,
    state: row.state,
    latitude: toFloat(row.lat),
    longitude: toFloat(row.long),
    cityPopulation: toInteger(row.city_pop),
    isFraud: row.is_fraud = '1'
})

// Create the relationship between Transactions and Merchants
MERGE (m)<-[:CONDUCTED_AT]-(t);