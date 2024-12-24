package dev.hungq.movie_service.movie;

import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;

@Service
public class MovieService {

	private final MovieRepository movieRepo;
	
	public MovieService(MovieRepository movieRepo)
	{
		this.movieRepo = movieRepo;
	}

    public List<Movie> findAll() {
        return movieRepo.findAll();
    }

    public Optional<Movie> find(Integer id) {
        return movieRepo.findById(id);
    }

    public Movie save(Movie movie) {
        return movieRepo.save(movie);
    }

    public void delete(Integer id) {
    	movieRepo.deleteById(id);
    }
}