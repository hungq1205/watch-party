package dev.hungq.movie_service.movie;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;

@Entity
public class Movie {
	@Id
	@GeneratedValue(strategy=GenerationType.IDENTITY)
	private Integer id;
	private String title;
	private String url;
	@Column(name = "poster_url")
	private String posterUrl;

	public Movie() { }
	
	public Movie(String title, String url, String posterUrl) {
		super();
		this.title = title;
		this.url = url;
		this.posterUrl = posterUrl;
	}
	
	public Integer getId() {
		return id;
	}
	public void setId(Integer id) {
		this.id = id;
	}
	public String getTitle() {
		return title;
	}
	public void setTitle(String title) {
		this.title = title;
	}
	public String getUrl() {
		return url;
	}
	public void setUrl(String url) {
		this.url = url;
	}
	public String getPosterUrl() {
		return posterUrl;
	}
	public void setPosterUrl(String posterUrl) {
		this.posterUrl = posterUrl;
	}
	
	@Override
	public String toString() {
		return "Movie [id=" + id + ", title=" + title + ", url=" + url + ", posterUrl=" + posterUrl + "]";
	}
}
