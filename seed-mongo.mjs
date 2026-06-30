import { MongoClient } from "mongodb";

const uri = "mongodb://mongo:cCfdrXNwHYjPPTAEFUoizdnwEDzhdomz@reseau.proxy.rlwy.net:23192";
const client = new MongoClient(uri);
await client.connect();
const db = client.db("calli_machtia");

await db.createCollection("modules");
await db.collection("modules").createIndex({ course_id: 1 });
await db.collection("modules").createIndex({ course_id: 1, order: 1 });

console.log("MongoDB collections and indexes created successfully");
await client.close();
