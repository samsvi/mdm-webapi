const mongoHost = process.env.MDM_API_MONGODB_HOST;
const mongoPort = process.env.MDM_API_MONGODB_PORT;

const mongoUser = process.env.MDM_API_MONGODB_USERNAME;
const mongoPassword = process.env.MDM_API_MONGODB_PASSWORD;

const database =
  process.env.MDM_API_MONGODB_DATABASE || "mdm-patient-management";
const patientsCollection = "patients";
const medicalRecordsCollection = "medical-records";

const retrySeconds = parseInt(process.env.RETRY_CONNECTION_SECONDS || "5") || 5;

// try to connect to mongoDB until it is not available
let connection;
while (true) {
  try {
    connection = Mongo(
      `mongodb://${mongoUser}:${mongoPassword}@${mongoHost}:${mongoPort}`
    );
    break;
  } catch (exception) {
    print(`Cannot connect to mongoDB: ${exception}`);
    print(`Will retry after ${retrySeconds} seconds`);
    sleep(retrySeconds * 1000);
  }
}

// if database and collections exist, exit with success - already initialized
const databases = connection.getDBNames();
if (databases.includes(database)) {
  const dbInstance = connection.getDB(database);
  collections = dbInstance.getCollectionNames();
  if (
    collections.includes(patientsCollection) &&
    collections.includes(medicalRecordsCollection)
  ) {
    print(
      `Collections '${patientsCollection}' and '${medicalRecordsCollection}' already exist in database '${database}'`
    );
    process.exit(0);
  }
}

// initialize
// create database and collections
const db = connection.getDB(database);
db.createCollection(patientsCollection);
db.createCollection(medicalRecordsCollection);

// create indexes
db[patientsCollection].createIndex({ id: 1 });
db[patientsCollection].createIndex({ insuranceNumber: 1 });
db[medicalRecordsCollection].createIndex({ id: 1 });
db[medicalRecordsCollection].createIndex({ patientId: 1 });

//insert sample data - patients
let patientsResult = db[patientsCollection].insertMany([
  {
    id: "pat123456",
    firstName: "Ján",
    lastName: "Novák",
    dateOfBirth: "1990-01-01",
    gender: "M",
    insuranceNumber: "900101/1234",
    bloodType: "A+",
    status: "Stable",
    allergies: "Penicilín, arašidy",
    medicalNotes: "Pacient má chronické problémy s tlakom",
    createdAt: new Date(),
    updatedAt: new Date(),
  },
  {
    id: "pat789012",
    firstName: "Anna",
    lastName: "Svobodová",
    dateOfBirth: "1985-03-15",
    gender: "F",
    insuranceNumber: "850315/5678",
    bloodType: "O-",
    status: "Recovering",
    allergies: "",
    medicalNotes: "",
    createdAt: new Date(),
    updatedAt: new Date(),
  },
]);

if (patientsResult.writeError) {
  console.error(patientsResult);
  print(`Error when writing patients data: ${patientsResult.errmsg}`);
}

//insert sample data - medical records
let recordsResult = db[medicalRecordsCollection].insertMany([
  {
    id: "rec789012",
    patientId: "pat123456",
    dateOfVisit: new Date("2024-05-15T09:30:00Z"),
    diagnosis: "Akútna respiračná infekcia",
    symptoms: ["kašeľ", "teploty", "bolesti hrdla"],
    treatment: "Predpísané antibiotiká, odpočinok, zvýšený príjem tekutín",
    medications: [
      {
        name: "Amoxicillin",
        dosage: "500mg",
        frequency: "3x denne",
        duration: "7 dní",
      },
    ],
    doctorName: "Dr. Peter Kováč",
    notes: "Pacient má alergiu na penicilín",
    followUpDate: "2024-05-22",
    createdAt: new Date("2024-05-15T09:30:00Z"),
    updatedAt: new Date("2024-05-15T09:30:00Z"),
  },
  {
    id: "rec789013",
    patientId: "pat123456",
    dateOfVisit: new Date("2024-03-10T14:00:00Z"),
    diagnosis: "Preventívna prehliadka",
    symptoms: [],
    treatment: "Kontrola zdravotného stavu",
    medications: [],
    doctorName: "Dr. Eva Horáková",
    notes: "Všetko v poriadku",
    followUpDate: "2025-03-10",
    createdAt: new Date("2024-03-10T14:00:00Z"),
    updatedAt: new Date("2024-03-10T14:00:00Z"),
  },
]);

if (recordsResult.writeError) {
  console.error(recordsResult);
  print(`Error when writing medical records data: ${recordsResult.errmsg}`);
}

print(
  `Successfully initialized MDM database with ${patientsResult.insertedIds.length} patients and ${recordsResult.insertedIds.length} medical records`
);

// exit with success
process.exit(0);
