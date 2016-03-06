
  jQuery.ajaxSetup({
    "contentType": "application/json; charset=UTF-8"
  });

  function setCanvasSize(factor) {
    console.log("sizing canvas");
    var canvas = document.getElementById('mynetwork');
    var relativ = document.getElementById('canvassizer');
    console.log({"canvas": canvas, "relativ": relativ});
    canvas.width = factor * jQuery("#canvassizer").width();
    canvas.height = factor * jQuery("#canvassizer").height();
  }

 setCanvasSize(1);

/*
  // create an array with nodes
  var nodes = new vis.DataSet([
    {id: "a", "label": "universalsprache", "value": 0, "title": "ideen, IT"},
    {id: "b", "label": "künstliche intelligenz", "value": 1, "title": "ideen, IT"},
    {id: "c", "label": "der Riese", "value": 2, "title": "zufällig"},
    {id: "d", "label": "unverbunden", "value": 0},
    {id: "e", "label": "ein bisschen", "value": 1, "title": "zufällig"}
  ]);

  // create an array with edges
  var edges = new vis.DataSet([
    {from: "a", to: "c"},
    {from: "a", to: "b"},
    {from: "b", to: "c"},
    {from: "b", to: "e"}
  ]);
  */

  // create a network
  var container = document.getElementById('mynetwork');
 /*
  var data = {
    nodes: nodes,
    edges: edges
  };
  */
  /* locales: locales, */
  var options = {
    autoResize: true,
    height: '100%',
    width: '100%',
    locale: 'en',
    clickToUse: false,
    
    configure: {
      enabled: false
    },
    
    edges: {
      arrows: {
        to: {
          enabled: true,
          scaleFactor: 1
        }
      }
    },
    nodes: {
      shape: "box",
      scaling: {
        label: {
          enabled: true
        }
      }
    },
    layout: {
      hierarchical: {
        enabled: true,
        direction: "DU",
        sortMethod: "directed"
      }
    },
    manipulation: {
      enabled: true,
      initiallyActive: true,
      addNode: function(nodeData,callback) {
        var newname = prompt("Enter name for new node:", "");
        if (newname) {
          nodeData.label = newname;
          jQuery.ajax({
            method: "PUT",
            url: "/item/put",
            data: JSON.stringify({
              "Name": newname
            }),
            contentType: "application/json; charset=UTF-8",
            success: function(){ 
              callback(nodeData);
              getData(); }
          });
        } else {
          callback();
        }
      },
      addEdge: function(edgeData,callback) {
        // no backlinks
        if (edgeData.from === edgeData.to) {
          callback();
          return
        }
        var fromName = network.body.nodes[edgeData.from].labelModule.nodeOptions.label;
        var toName = network.body.nodes[edgeData.to].labelModule.nodeOptions.label;

        jQuery.ajax({
          method: "PUT",
          url: "/item/put-edge",
          data: JSON.stringify({
            "From": fromName,
            "To": toName,
          }),
          contentType: "application/json; charset=UTF-8",
          success: function(){ 
            callback(edgeData);
            getData(); }
        });
      },
      deleteNode: function(deleteArr, callback) {
        //console.log(deleteArr.nodes[0]);
        var nodeName = network.body.nodes[deleteArr.nodes[0]].labelModule.nodeOptions.label;
        jQuery.ajax({
          method: "DELETE",
          url: "/item/remove",
          data: JSON.stringify({
            "Name": nodeName
          }),
          contentType: "application/json; charset=UTF-8",
          success: function(){ 
            callback();
            getData(); }
        });
      },
      deleteEdge: function(deleteArr, callback) {
        // console.log(deleteArr);
        var edge = network.body.edges[deleteArr.edges[0]];
        // console.log(edge);
        var nodeFrom = network.body.nodes[edge.fromId].labelModule.nodeOptions.label;
        var nodeTo = network.body.nodes[edge.toId].labelModule.nodeOptions.label;
        jQuery.ajax({
          method: "DELETE",
          url: "/item/remove-edge",
          data: JSON.stringify({
            "From": nodeFrom,
            "To": nodeTo
          }),
          contentType: "application/json; charset=UTF-8",
          success: function(){ 
            callback();
            getData(); }
        });
      },
      /*
      editEdge: function...
      */
      editNode: function(nodeData,callback) {
        var newname = prompt("Enter new name for node:", nodeData.label);
        if (newname && newname != nodeData.label) {
          jQuery.ajax({
            method: "PATCH",
            url: "/item/rename",
            data: JSON.stringify({
              "Old": nodeData.label,
              "New": newname,
            }),
            contentType: "application/json; charset=UTF-8",
            success: function(){ 
              nodeData.label = newname;
              callback(nodeData);
              getData(); }
          });
        } else {
          callback();
        }
      }
    },
    interaction: {
      dragNodes: false,
      keyboard: {
        enabled: true,
        bindToWindow: true
      }
    },
    groups:{
      useDefaultGroups: true,
      group0:{
        color: "lightgray"
      },
      group1:{
        color: "black",
        font: {
          color: "white"
        } 
      },
      group2:{
        color: "green"
      },
      group3:{
        color: "yellow"
      },
      group4:{
        color: "blue",
        font: {
          color: "white"
        } 
      },
      group5:{
        color: "red",
        font: {
          color: "white"
        } 
      }
    }
  };
  /*,
  configure: {...},    // defined in the configure module.
  edges: {...},        // defined in the edges module.
  nodes: {...},        // defined in the nodes module.
  groups: {...},       // defined in the groups module.
  layout: {...},       // defined in the layout module.
  interaction: {...},  // defined in the interaction module.
  manipulation: {...}, // defined in the manipulation module.
  physics: {...},      // defined in the physics module.
  */
  var network = new vis.Network(container, {nodes: [], edges: []}, options);

  function getData(callback) {
    jQuery.getJSON("/item/vis", function(data){
      console.log(data);
      network.setData(data);
      network.redraw();  
      if (callback) {
        callback();
      }
    });
  }

  getData();

  jQuery.getJSON("/app/name", function(data) {
    document.title = data.Name + " | prioritize";
  })
  /*
  getData(function(data) {
    console.log(data);
    network.setData(data);
    network.redraw();
  });
  */

  // loads new data
  // network.setData(data);

  // set new options
// network.setOptions(options);

// redraw
// network.redraw()  