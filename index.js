const sql = require("sqlite3");

const db = new sql.Database("./main.db");

async function main() {
  try {
    const gates = await new Promise(function (resolve, reject) {
      db.all(
        `SELECT * FROM products WHERE type = 'gate'`,
        function (err, rows) {
          if (err) return reject(err);
          resolve(rows);
        }
      );
    });
    const extensions = await new Promise(function(resolve, reject){
        db.all(`SELECT * FROM products WHERE type = 'extension'`, function(err, rows){
            if (err) return reject(err)
            resolve(rows)
        })
    })
    for(const gate of gates){
        for(const extension of extensions) {
            if(gate.color === extension.color){
                await new Promise(function(resolve, reject){
                    db.run(`INSERT INTO compatibles (gate_id, extension_id) VALUES (?, ?)`, [gate.id, extension.id], function(err){
                        if (err) return reject(err)
                        resolve()
                    })
                })
            }
        }
    }
  } catch (err) {
    console.log(err);
  }
}


main()