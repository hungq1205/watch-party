package dev.hungq.movie_service.box;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/box")
public class BoxController {
	
	private final BoxService boxService; 
	
	public BoxController(BoxService boxService)
	{
		this.boxService = boxService;
	}
	
	@GetMapping("/all")
	ResponseEntity<List<Box>> getBoxes() 
	{
        return new ResponseEntity<List<Box>>(boxService.findAll(), HttpStatus.OK);
	}

	@GetMapping("/{id}")
	ResponseEntity<Box> getBox(@PathVariable int id)
	{
		var b = boxService.find(id);
		if (b.isEmpty())
			return new ResponseEntity<Box>(HttpStatus.NOT_FOUND);
        return new ResponseEntity<Box>(boxService.save(b.get()), HttpStatus.OK);
	}

	@GetMapping("/owner/{id}")
	ResponseEntity<Box> getBoxOfOwner(@PathVariable int id)
	{
		var b = boxService.findByOwnerId(id);
		if (b.isEmpty())
			return new ResponseEntity<Box>(HttpStatus.NOT_FOUND);
        return new ResponseEntity<Box>(boxService.save(b.get()), HttpStatus.OK);
	}

	@GetMapping("/user/{id}")
	ResponseEntity<Box> getBoxOfUser(@PathVariable int id)
	{
		var b = boxService.findByUserId(id);
		if (b.isEmpty())
			return new ResponseEntity<Box>(HttpStatus.NOT_FOUND);
        return new ResponseEntity<Box>(b.get(), HttpStatus.OK);
	}

	@GetMapping("/{boxId}/exists/{userId}")
	ResponseEntity<Map<String, Boolean>> getBoxOfUser(@PathVariable int boxId, @PathVariable int userId)
	{
		var b = boxService.containsUser(boxId, userId);
	    Map<String, Boolean> response = new HashMap<>();
	    response.put("value", b);
	    return new ResponseEntity<>(response, HttpStatus.OK);
	}
	
	@PostMapping("")
	ResponseEntity<Box> createBox(@RequestBody Box box)
	{
        return new ResponseEntity<Box>(boxService.create(box), HttpStatus.OK);
	}

	@PutMapping("/{id}")
	ResponseEntity<Box> updateBox(@PathVariable int id, @RequestBody Box box)
	{
		var obox = boxService.find(id);
		
		if (obox.isEmpty())
			return new ResponseEntity<Box>(HttpStatus.NOT_FOUND);
		var b = obox.get();
		b.setOwnerId(box.getOwnerId());
		b.setElapsed(box.getElapsed());
		b.setMovieUrl(box.getMovieUrl());
		b.setPassword(box.getPassword());
		
        return new ResponseEntity<Box>(boxService.save(b), HttpStatus.OK);
	}

	@DeleteMapping("/{id}")
	ResponseEntity<String> deleteBox(@PathVariable int id)
	{
		boxService.delete(id);
		return new ResponseEntity<String>("Box " + id + " deleted", HttpStatus.OK);
	}
	
	@DeleteMapping("")
	ResponseEntity<String> deleteOfOwner(@RequestParam(name = "owner_id", required=true) int ownerId)
	{
		boxService.deleteByOwnerId(ownerId);
		return new ResponseEntity<String>("Box of owner " + ownerId + " deleted", HttpStatus.OK);
	}
	
	@PostMapping("/{boxId}/add/{userId}")
	ResponseEntity<String> addUserToBox(@PathVariable Integer boxId, @PathVariable Integer userId) 
	{
		if (boxService.addUserToBox(boxId, userId))
			return new ResponseEntity<String>("Added user " + userId + " to box " + boxId, HttpStatus.OK);
		return new ResponseEntity<String>("Failed to add user " + userId + " to box " + boxId, HttpStatus.BAD_REQUEST);
	}
	
	@DeleteMapping("/{boxId}/remove/{userId}")
	ResponseEntity<String> removeUserFromBox(@PathVariable Integer boxId, @PathVariable Integer userId) 
	{
		if (boxService.removeUserFromBox(boxId, userId))
			return new ResponseEntity<String>("Removed user " + userId + " from box " + boxId, HttpStatus.OK);
		return new ResponseEntity<String>("Failed to remove user " + userId + " from box " + boxId, HttpStatus.BAD_REQUEST);
	}
}
