/*

The Project is divided in two main parts:
- Public APIs
- Internal APIs

The Public APIs have the only job of fetching the recommendations from the database given a specified model in the request body.
For more information about "How to create the request", check the endpoint "/docs". A swagger page should appear.

The Internal APIs do the hard working on handling the creation/deletion/update of models and data. The database we chose for serving
the recommendations is Aerospike. Aerospike is a Key/Value storage database similar to Redis but with steroids. It has a concept of
tables (and columns) which makes it easier to create a structured definition of both models and data.

Leaving aside how Aerospike works (check the documentation for more information https://www.aerospike.com/docs/), it is important
to point out how we designed the internal system.
We have two concepts:
- model
- data

The model is a simple "table" in which contains the metadata of model itself. This is needed for then creating the appropriate Key in
the "data" table(s). The "data" is organized similarly but with different specifications.
In aerospike the hierarchy is the following (where "-->"" means "contains many uniques items of"):

Namespace --> setName --> Keys --> Bins


The Models

Namespace: phoenix
SetName: publicationPoint
Key: Campaign
Bins: version => 0.1.0 				 // as start
	  stage => STAGED/PUBLISHED		 // either value
	  signalType => articleID_userID // this is an example

The namespace is the main "container" for all the data we want to store. Hence, it is common for both concepts.
The setname (you can think this as the table in RDBMS) corresponds to the "publicationPoint". Since we can have multiple campaigns for
the same publicationPoint, the Key becomes the campaign. The Bins are the "values" of the "Key". Each Bin has also a "key/value" pair
and the entries are specified in the schema above.
Every time an action is done on the model (publish the model, stage the model, etc), the Version is either increased/decreased based
on the SemVer algorithm.

Below you can find an example of multiple models:

- Model1
	SetName = rtl_news
	Key = homepage
- Model2
	SetName = rtl_news
	Key = footer
- Model3
	SetName = videoland
	Key = profile


The Data

Namespace: phoenix
SetName: publicationPoint#campaign#STAGED/PUBLISHED
Key: signalID	// for example 111_3333
Bins: signalID => ["item1", "item2", ..., "itemN"]

The data is organized similarly to the Model but with a different naming convention. To make the SetName "unique" per model, we
use a combination of "publicPoint", "campaing" and "stage" separated by a #. In this way, we are able to insert all the Keys we
need for that particular model

Below you can find and example of data for a model

- Data1
	SetName = rtl_news#homepage#STAGED
		- Key = 11_22
		  Bins = 11_22 = ["1","2","3"]
		- Key = 33_44
		  Bins = 33_44 = ["4","5","6"]
		- Key = 55_66
		  Bins = 55_66 = ["7","8","9"]
- Data2
	SetName = rtl_news#footer#PUBLISHED
		- Key = 3333
		  Bins = 3333 = ["a","b","c"]
		- Key = 4444
		  Bins = 4444 = ["d","e","f"]
		- Key = 5555
		  Bins = 5555 = ["g","h","i"]
*/

package docs
