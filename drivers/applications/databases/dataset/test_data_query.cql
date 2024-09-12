LOAD CSV WITH HEADERS FROM 'file:///test_table_export.csv' AS row
           MERGE (n:Node {key: row.key, value: row.value})
           WITH n
           ORDER BY n.key
           WITH collect(n) as nodes
           FOREACH (i in RANGE(0, size(nodes)-2) |
             FOREACH (n1 in [nodes[i]] |
               FOREACH (n2 in [nodes[i+1]] |
                 MERGE (n1)-[:NEXT]->(n2)
               )
             )
           );