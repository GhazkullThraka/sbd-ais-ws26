# Exercise 2 - Solutions

## 2.1:
```bash
DROP TABLE IF EXISTS orders;

CREATE TABLE orders (
  customer_name TEXT,
  product_category TEXT,
  quantity INTEGER,
  price_per_unit NUMERIC,
  order_date timestamp,
  country TEXT
);

\COPY orders(customer_name,product_category,quantity,price_per_unit,order_date,country) FROM '/data/orders_1M.csv' DELIMITER ',' CSV HEADER;
```
### A:  Which order has the highest price_per_unit?
```bash
SELECT * 
FROM orders 
ORDER BY price_per_unit DESC;
```
Automotive order from Emma Brown, PPU: 2000

### B: What are the top 3 product categories with the highest total quantity sold across all orders?
```bash
SELECT product_category, SUM(quantity)
FROM orders
GROUP BY product_category
ORDER BY SUM(quantity) DESC
```
1. Health and Beauty
2. Electronics
3. Toys

### C: What is the total revenue per product category? (Revenue = price_per_unit × quantity)
```bash
SELECT product_category, SUM(price_per_unit * quantity) AS revenue
FROM orders
GROUP BY product_category
ORDER BY revenue DESC;
```
Automotive       | 306589798.86<br>
Electronics      | 241525009.45<br>
Home & Garden    |  78023780.09<br>

### D. Who are the top 5 customers by total spending?
```bash
SELECT customer_name, SUM(price_per_unit * quantity) AS revenue
FROM orders
GROUP BY customer_name
ORDER BY revenue DESC;
```
Carol Taylor        | 991179.18<br>
Nina Lopez          | 975444.95<br>
Daniel Jackson      | 959344.48<br>
Carol Lewis         | 947708.57<br>
Daniel Young        | 946030.14<br>

### E. Look at the spending totals in D — and at how many orders each of those customers has.
```bash
SELECT customer_name, SUM(price_per_unit * quantity) AS revenue, COUNT(*)
FROM orders
GROUP BY customer_name
ORDER BY revenue DESC;
```
The counts for the top 5 biggest spenders are in the same ballpark as every other customer.
Looking at dataset_generator.py it's because there are only so many first and last names. Therefore, for 1M orders each unique combination only gets a certain number of orders on average.

## 2.2
it would grow by k^2<br><b>Prediction:</b>

| rows in table | join result (`COUNT(*)`) | time |
|---|---|---|
| 50 000 | 2.5b| 10m |
| 100 000 | 10b| 30m |
| 200 000 |40b | 60m |

<b>Actual:</b>

| rows in table | join result (`COUNT(*)`) | time |
|---|---|---|
| 50,000 | 27,501,822 | Time: 3542.463 ms (00:03.542)|
| 100,000 | 109,946,508 |Time: 15590.393 ms (00:15.590) |
| 200,000 | 439,395,606 | Time: 89882.683 ms (01:29.883)S|

```bash
CREATE INDEX countries_index
on people_100k (country)

ANALYZE people_100k;

EXPLAIN ANALYZE
SELECT COUNT(*)
FROM people_100k p1
JOIN people_100k p2
  ON p1.country = p2.country;
```
Execution Time: 33668.173 ms
```bash
 Finalize Aggregate  (cost=641922.22..641922.23 rows=1 width=8) (actual time=33647.169..33666.243 rows=1.00 loops=1)
   Buffers: shared hit=2048
   ->  Gather  (cost=641922.00..641922.21 rows=2 width=8) (actual time=33637.089..33666.185 rows=2.00 loops=1)
         Workers Planned: 1
         Workers Launched: 1
         Buffers: shared hit=2048
         ->  Partial Aggregate  (cost=640922.00..640922.01 rows=1 width=8) (actual time=33611.112..33611.119 rows=1.00 loops=2)
               Buffers: shared hit=2048
               ->  Parallel Hash Join  (cost=2347.54..478869.52 rows=64820993 width=0) (actual time=679.742..24677.931 rows=54973254.00 loops=2)
                     Hash Cond: (p1.country = p2.country)
                     Buffers: shared hit=2048
                     ->  Parallel Seq Scan on people_100k p1  (cost=0.00..1612.24 rows=58824 width=8) (actual time=0.012..29.715 rows=50000.00 loops=2)
                           Buffers: shared hit=1024
                     ->  Parallel Hash  (cost=1612.24..1612.24 rows=58824 width=8) (actual time=677.135..677.137 rows=50000.00 loops=2)
                           Buckets: 131072  Batches: 1  Memory Usage: 5248kB
                           Buffers: shared hit=1024
                           ->  Parallel Seq Scan on people_100k p2  (cost=0.00..1612.24 rows=58824 width=8) (actual time=551.942..559.867 rows=50000.00 loops=2)
                                 Buffers: shared hit=1024
 Planning Time: 1.051 ms
 JIT:
   Functions: 24
   Options: Inlining true, Optimization true, Expressions true, Deforming true
   Timing: Generation 3.153 ms (Deform 1.013 ms), Inlining 308.889 ms, Optimization 472.173 ms, Emission 323.005 ms, Total 1107.221 ms
 Execution Time: 33668.173 ms
(24 rows)
```
Despite the index the query plan still used a sequential scan. I think it didn't matter anyway since the query has to join the entire table anyway. Additionally, the join has to combine thousands of entries resulting in millions of pairs, an index does not address this at all.
## Rewrite
If a country has k people, then it will have k^2 for the join. We can optimize this directly.
```bash
SELECT SUM(k * k) 
FROM (
  SELECT COUNT(*) AS k
  FROM people_100k
  GROUP BY country
) AS counts;
```
people_100k:
109,946,508 | Time: 49.124 ms<br>
people_big:  10,983,941,260 | Time: 1590.679 ms (00:01.591)
## Discussion

1.  The rewrite demonstrates how having an efficient algorithm is much better than throwing more hardware at a problem. And an index doesn't help for a problem like this because the query has to go through every single row regardless.
2. You'd definitely need a better machine or bigger server cluster, however even giving it 4x the computing power would only allow it to process a dataset twice as large.
3. Normal OLTP databases are built for routine operations like updating a profile or checking a balance. Doing large and resource-intensive queries like a self join of this scale will quickly eat up all the resources of a live database. For a transaction load like this you'd be better off using an actual dedicated data warehouse.   


