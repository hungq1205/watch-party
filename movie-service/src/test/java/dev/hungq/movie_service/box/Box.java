package dev.hungq.movie_service.box;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;

import dev.hungq.movie_service.movie.Movie;
import jakarta.persistence.CollectionTable;
import jakarta.persistence.Column;
import jakarta.persistence.ElementCollection;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;

@Entity
public class Box {
	@Id 
	@GeneratedValue(strategy=GenerationType.IDENTITY)
	private Integer id;
	@Column(name = "owner_id")
	private Integer ownerId;
	@Column(name = "msg_box_id")
	private Integer msgBoxId;
	@Column(name = "elapsed")
	private Float elapsed;
	@Column(name = "password")
	private String password;
	
	@ManyToOne
    @JoinColumn(name = "movie_id")
    private Movie movie;
	
	@ElementCollection
	@CollectionTable(name = "box_user")
    @Column(name = "user_id")
	private List<Integer> userIds = new ArrayList<>();
	
	public Box() {}
	
	public Box(Integer ownerId, String password) {
		super();
		this.ownerId = ownerId;
		this.password = password;
		userIds.add(ownerId);
    }
	
	public Integer getId() {
		return id;
	}

	public void setId(Integer id) {
		this.id = id;
	}

	public Integer getOwnerId() {
		return ownerId;
	}

	public void setOwnerId(Integer ownerId) {
		this.ownerId = ownerId;
	}

	public Integer getMsgBoxId() {
		return msgBoxId;
	}

	public void setMsgBoxId(Integer msgBoxId) {
		this.msgBoxId = msgBoxId;
	}

	public Float getElapsed() {
		return elapsed;
	}

	public void setElapsed(Float elapsed) {
		this.elapsed = elapsed;
	}

    public Movie getMovie() {
        return movie;
    }

    public void setMovie(Movie movie) {
        this.movie = movie;
    }

	public String getPassword() {
		return password;
	}

	public void setPassword(String password) {
		this.password = password;
	}
	
	public List<Integer> getUserIds() {
		return userIds;
	}
	
	@JsonProperty("movie_id")
    public Integer getMovieId() {
        return (movie == null) ? -1 : movie.getId();
    }

	@Override
	public String toString() {
		return "Box [id=" + id + ", ownerId=" + ownerId + ", msgBoxId=" + msgBoxId + ", elapsed=" + elapsed
				+ ", movie=" + movie.getTitle() + " (" + movie.getId() + ")" + ", password=" + password + "]";
	}
}