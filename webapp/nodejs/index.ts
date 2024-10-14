// Get the client
import mysql, { Connection } from 'mysql2/promise';
import fs from 'fs';
import { off } from 'process';

// Create the connection to database
const connection = (await mysql.createConnection({
  host: 'localhost',
  user: 'isuconp',
  database: 'isuconp',
  password: 'isuconp',
})) ;

// A simple SELECT query
try {
  let pos = 0;
  while (true) {
  const [results] = await connection.query(
    `SELECT id, mime, imgdata FROM posts ORDER BY id LIMIT 10 OFFSET ${pos} `
  );
  pos += (results as any).length;
  if ((results as any).length === 0) {
    break;
  }
  for (const row of results as any) {
      console.log(row.id, row.mime);
      const ext = {
        'image/jpeg': 'jpg',
        'image/png': 'png',
        'image/gif': 'gif',
      }[row.mime];
      const path = `images/${row.id}.${ext}`;
      fs.writeFileSync(path, row.imgdata);
  }
}

  await connection.end();
} catch (err) {
  console.log(err);
}
