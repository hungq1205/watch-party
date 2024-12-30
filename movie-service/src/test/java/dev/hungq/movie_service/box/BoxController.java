package dev.hungq.movie_service.box;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

import dev.hungq.movie_service.movie.MovieService;

@RestController
@RequestMapping("/api/box")
public class BoxController {

	private final BoxService boxService; 
	private final MovieService movieService; 
	
	public BoxController(BoxService boxService, MovieService movieService)
	{
		this.boxService = boxService;
		this.movieService = movieService;
	}
	
	@GetMapping("")
	ResponseEntity<List<Box>> getBoxes() 
	{
        return new ResponseEntity<List<Box>>(boxService.findAll(), HttpStatus.OK);
	}

	@GetMapping("/{id}")
	ResponseEntity<Box> getBox(@PathVariable int id)
	{
		var b = boxService.find(id);
		if (b.isEmpty())
			return ResponseEntity.notFound().build();
	    return ResponseEntity.ok(b.get());
	}

	@GetMapping("/owner/{id}")
	ResponseEntity<Box> getBoxOfOwner(@PathVariable int id)
	{
		var b = boxService.findByOwnerId(id);
		if (b.isEmpty())
			return ResponseEntity.notFound().build();
        return ResponseEntity.ok(b.get());
	}

	@GetMapping("/user/{id}")
	ResponseEntity<Box> getBoxOfUser(@PathVariable int id)
	{
		var b = boxService.findByUserId(id);
		if (b.isEmpty())
			return ResponseEntity.notFound().build();
        return ResponseEntity.ok(b.get());
	}

	@GetMapping("/{boxId}/exists/{userId}")
	ResponseEntity<Map<String, Boolean>> existsUser(@PathVariable int boxId, @PathVariable int userId)
	{
		var b = boxService.containsUser(boxId, userId);
	    Map<String, Boolean> response = new HashMap<>();
	    response.put("value", b);
	    return ResponseEntity.ok(response);
	}
	
	@PostMapping(""	)
	ResponseEntity<Box> createBox(@RequestBody Box box)
	{
        return ResponseEntity.ok(boxService.create(box));
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
		b.setMovie(box.getMovie());
		b.setPassword(box.getPassword());
		
        return ResponseEntity.ok(boxService.save(b));
	}
	
	@PatchMapping("/{id}/movie")
	ResponseEntity<Box> updateMovieInBox(@PathVariable int id, @RequestBody Map<String, Integer> movieUpdate) {
	    var obox = boxService.find(id);

	    if (obox.isEmpty()) {
	        return ResponseEntity.notFound().build();
	    }

	    var box = obox.get();

	    Integer movieId = movieUpdate.get("movie_id");
	    if (movieId == null || movieId < 0) {
		    box.setMovie(null);
		    box.setElapsed(movieUpdate.get("elapsed").floatValue());
		    return ResponseEntity.ok(boxService.save(box));
	    }
	    var movie = movieService.find(movieId)
	    		.orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "Movie not found"));
	    
	    box.setMovie(movie);
	    box.setElapsed(movieUpdate.get("elapsed").floatValue());

	    return ResponseEntity.ok(boxService.save(box));
	}

	@DeleteMapping("/{id}")
	ResponseEntity<String> deleteBox(@PathVariable int id)
	{
		boxService.delete(id);
		return ResponseEntity.ok("Box " + id + " deleted");
	}
	
	@DeleteMapping("")
	ResponseEntity<String> deleteOfOwner(@RequestParam(name = "owner_id", required=true) int ownerId)
	{
		boxService.deleteByOwnerId(ownerId);
		return ResponseEntity.ok("Box of owner " + ownerId + " deleted");
	}
	
	@PostMapping("/{boxId}/add/{userId}")
	ResponseEntity<String> addUserToBox(@PathVariable Integer boxId, @PathVariable Integer userId) 
	{
		if (boxService.addUserToBox(boxId, userId))
			return ResponseEntity.ok("Added user " + userId + " to box " + boxId);
		return new ResponseEntity<String>("Failed to add user " + userId + " to box " + boxId, HttpStatus.BAD_REQUEST);
	}
	
	@DeleteMapping("/{boxId}/remove/{userId}")
	ResponseEntity<String> removeUserFromBox(@PathVariable Integer boxId, @PathVariable Integer userId) 
	{
		if (boxService.removeUserFromBox(boxId, userId))
			return ResponseEntity.ok("Removed user " + userId + " from box " + boxId);
		return new ResponseEntity<String>("Failed to remove user " + userId + " from box " + boxId, HttpStatus.BAD_REQUEST);
	}
}
