package dev.hungq.movie_service.movie;

import java.util.List;

import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

@RestController
@RequestMapping("/api/movie")
public class MovieController {
	
	private final MovieService movieService; 
	
	public MovieController(MovieService movieService)
	{
		this.movieService = movieService;
	}
	
	@GetMapping
	List<Movie> getMovies() 
	{
		return movieService.findAll();
	}
	
	@GetMapping("/{id}")
	Movie getMovie(@PathVariable int id)
	{
		var m = movieService.find(id);
		if (m.isEmpty())
			throw new ResponseStatusException(HttpStatus.NOT_FOUND);
		
		return m.get();
	}
}
