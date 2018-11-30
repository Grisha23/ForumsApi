

WITH recursive post_tree(id,path,author,created,forum,isedited,message,parent,thread) AS (
  SELECT p.id, array_append('{}'::bigint[], id) AS arr_id, p.author,p.created,p.forum,p.isedited,p.message,p.parent,p.thread
    FROM posts p
      WHERE p.parent = 0 AND p.thread=3

  UNION ALL

  SELECT p.id, array_append(path, p.id), p.author,p.created,p.forum,p.isedited,p.message,p.parent,p.thread
    FROM posts p
      JOIN post_tree pt
        ON p.parent = pt.id
)
SELECT p.author,p.created,p.forum,p.id,p.parent,p.thread,path from post_tree p ORDER by path[0],path LIMIT 65;

SELECT p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread
  from post_tree pt
    JOIN posts p on p.id = pt.id;
      where path > (select path from post_tree where id = SINCE VAL)
         order by path[0], path ASC-- " + sinceAddition + " " + sortAddition + " " + limitAddition;





WITH recursive post_tree(id,path) AS (SELECT p.id, array_append('{}'::bigint[], id) AS arr_id FROM posts p WHERE p.parent=0 AND p.thread=2 UNION ALL SELECT p.id, array_append(path, p.id) FROM posts p JOIN post_tree pt ON p.parent = pt.id)
select p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread from post_tree pt join posts p on p.id = pt.id;

------------------------------------------------------------------------------------
select author,created,forum,id,isedited,message,parent,thread
  from
    (
      WITH recursive post_tree(id,path,author,created,forum,isedited,message,parent,thread) as
      (
	      select p.id,array_append('{}'::bigint[], p.id) as arr_id , p.author,p.created,p.forum,p.isedited,p.message,p.parent,p.thread
	       from posts p
	         where p.parent = 0 and p.thread = $1

	    union all

	    select p.id, array_append(path, p.id), p.author,p.created,p.forum,p.isedited,p.message,p.parent,p.thread
	      from posts p
	        join post_tree pt on p.parent = pt.id
	    )
	    select post_tree.id as id, path,author,created,forum,isedited,message,parent,thread,
	      dense_rank() over (order by path[1]   descflag   ) as r
	        from post_tree   sinceAddition
	  ) as pt  limitAddition  sortAddition  -- join posts p on p.id = pt.id  limitAddition  sortAddition